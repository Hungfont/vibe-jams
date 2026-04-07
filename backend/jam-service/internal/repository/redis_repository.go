package repository

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"sync"

	"video-streaming/backend/jams/internal/model"
)

var (
	// ErrQueueItemNotFound indicates remove/reorder references an unknown item.
	ErrQueueItemNotFound = errors.New("queue item not found")
	// ErrVersionConflict indicates reorder expected version is stale.
	ErrVersionConflict = errors.New("queue version conflict")
	// ErrIdempotencyKeyRequired indicates add missed idempotency key.
	ErrIdempotencyKeyRequired = errors.New("idempotency key required")
	// ErrSessionNotFound indicates requested session does not exist.
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionEnded indicates write command attempted against ended session.
	ErrSessionEnded = errors.New("session ended")
	// ErrHostOwnershipRequired indicates actor is not host for host-only operation.
	ErrHostOwnershipRequired = errors.New("host ownership required")
	// ErrParticipantNotFound indicates user is not in session participant list.
	ErrParticipantNotFound = errors.New("participant not found")
	// ErrModerationBlocked indicates actor is blocked by moderation policy.
	ErrModerationBlocked = errors.New("moderation blocked")
	// ErrModerationTargetInvalid indicates moderation target is invalid.
	ErrModerationTargetInvalid = errors.New("invalid moderation target")
)

// VersionConflictError captures stale optimistic concurrency metadata.
type VersionConflictError struct {
	ExpectedQueueVersion int64
	CurrentQueueVersion  int64
}

func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("queue version conflict: expected %d current %d", e.ExpectedQueueVersion, e.CurrentQueueVersion)
}

// Is allows errors.Is(err, ErrVersionConflict) checks.
func (e *VersionConflictError) Is(target error) bool {
	return target == ErrVersionConflict
}

// VersionConflictCurrentQueueVersion extracts authoritative queue version from conflict error.
func VersionConflictCurrentQueueVersion(err error) (int64, bool) {
	var conflictErr *VersionConflictError
	if !errors.As(err, &conflictErr) {
		return 0, false
	}

	return conflictErr.CurrentQueueVersion, true
}

type addResult struct {
	snapshot model.QueueSnapshot
}

type jamQueueState struct {
	items             []model.QueueItem
	queueVersion      int64
	nextItemSequence  int64
	idempotencyResult map[string]addResult
	sessionVersion    int64
	status            model.SessionStatus
	hostUserID        string
	participants      map[string]model.SessionRole
	mutedUsers        map[string]bool
	kickedUsers       map[string]bool
	endCause          string
	endedBy           string
}

// RedisQueueRepository stores queue state using in-memory maps with Redis-like key model.
type RedisQueueRepository struct {
	mu                  sync.Mutex
	jams                map[string]*jamQueueState
	nextSessionSequence int64
	storagePath         string
}

// NewRedisQueueRepository builds a repository implementation used by queue service.
func NewRedisQueueRepository() *RedisQueueRepository {
	return &RedisQueueRepository{
		jams: make(map[string]*jamQueueState),
	}
}

// NewDurableQueueRepository builds a repository with file-backed persistence.
func NewDurableQueueRepository(storagePath string) (*RedisQueueRepository, error) {
	repo := &RedisQueueRepository{
		jams:        make(map[string]*jamQueueState),
		storagePath: storagePath,
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if err := repo.loadDurableStateLocked(); err != nil {
		return nil, err
	}

	return repo, nil
}

// QueueItemsKey returns Redis list key format for ordered queue items.
func QueueItemsKey(jamID string) string {
	return fmt.Sprintf("jams:%s:queue:items", jamID)
}

// QueueMetadataKey returns Redis hash key format for queue metadata/version.
func QueueMetadataKey(jamID string) string {
	return fmt.Sprintf("jams:%s:queue:meta", jamID)
}

// QueueIdempotencyKey returns Redis hash key format for idempotency values.
func QueueIdempotencyKey(jamID string) string {
	return fmt.Sprintf("jams:%s:queue:idempotency", jamID)
}

// SessionMetadataKey returns Redis hash key format for session metadata.
func SessionMetadataKey(jamID string) string {
	return fmt.Sprintf("jams:%s:session:meta", jamID)
}

// SessionMembersKey returns Redis hash key format for session members.
func SessionMembersKey(jamID string) string {
	return fmt.Sprintf("jams:%s:session:members", jamID)
}

// SessionPermissionsKey returns Redis hash key format for session permissions.
func SessionPermissionsKey(jamID string) string {
	return fmt.Sprintf("jams:%s:session:permissions", jamID)
}

// CreateSession creates a new jam session with requester as host.
func (r *RedisQueueRepository) CreateSession(hostUserID string) (model.SessionSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextSessionSequence++
	jamID := fmt.Sprintf("jam_%d", r.nextSessionSequence)

	state := &jamQueueState{
		items:             make([]model.QueueItem, 0),
		idempotencyResult: make(map[string]addResult),
		sessionVersion:    1,
		status:            model.SessionStatusActive,
		hostUserID:        hostUserID,
		participants: map[string]model.SessionRole{
			hostUserID: model.SessionRoleHost,
		},
		mutedUsers:  make(map[string]bool),
		kickedUsers: make(map[string]bool),
	}
	r.jams[jamID] = state
	if err := r.saveDurableStateLocked(); err != nil {
		return model.SessionSnapshot{}, err
	}
	return buildSessionSnapshot(jamID, state), nil
}

// JoinSession adds one participant to an active session.
func (r *RedisQueueRepository) JoinSession(jamID string, userID string) (model.SessionSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.getSessionState(jamID)
	if err != nil {
		return model.SessionSnapshot{}, err
	}
	if err := ensureActive(state); err != nil {
		return model.SessionSnapshot{}, err
	}
	if state.kickedUsers[userID] {
		return model.SessionSnapshot{}, ErrModerationBlocked
	}
	if _, exists := state.participants[userID]; !exists {
		state.participants[userID] = model.SessionRoleMember
		state.sessionVersion++
		if err := r.saveDurableStateLocked(); err != nil {
			return model.SessionSnapshot{}, err
		}
	}

	return buildSessionSnapshot(jamID, state), nil
}

// LeaveSession removes one participant from an active session.
// If the host leaves, the session is ended with cause host_leave.
func (r *RedisQueueRepository) LeaveSession(jamID string, userID string) (model.SessionSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.getSessionState(jamID)
	if err != nil {
		return model.SessionSnapshot{}, err
	}
	if err := ensureActive(state); err != nil {
		return model.SessionSnapshot{}, err
	}
	role, exists := state.participants[userID]
	if !exists {
		return model.SessionSnapshot{}, ErrParticipantNotFound
	}

	if role == model.SessionRoleHost {
		state.status = model.SessionStatusEnded
		state.endCause = "host_leave"
		state.endedBy = userID
		state.sessionVersion++
		if err := r.saveDurableStateLocked(); err != nil {
			return model.SessionSnapshot{}, err
		}
		return buildSessionSnapshot(jamID, state), nil
	}

	delete(state.participants, userID)
	delete(state.mutedUsers, userID)
	state.sessionVersion++
	if err := r.saveDurableStateLocked(); err != nil {
		return model.SessionSnapshot{}, err
	}
	return buildSessionSnapshot(jamID, state), nil
}

// MuteParticipant marks one target participant as muted.
func (r *RedisQueueRepository) MuteParticipant(jamID string, actorUserID string, req model.ModerationCommandRequest) (model.SessionSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.getSessionState(jamID)
	if err != nil {
		return model.SessionSnapshot{}, err
	}
	if err := ensureActive(state); err != nil {
		return model.SessionSnapshot{}, err
	}
	if state.hostUserID != actorUserID {
		return model.SessionSnapshot{}, ErrHostOwnershipRequired
	}
	if req.TargetUserID == "" || req.TargetUserID == state.hostUserID {
		return model.SessionSnapshot{}, ErrModerationTargetInvalid
	}
	if _, exists := state.participants[req.TargetUserID]; !exists {
		return model.SessionSnapshot{}, ErrParticipantNotFound
	}

	state.mutedUsers[req.TargetUserID] = true
	state.sessionVersion++
	if err := r.saveDurableStateLocked(); err != nil {
		return model.SessionSnapshot{}, err
	}
	return buildSessionSnapshot(jamID, state), nil
}

// KickParticipant removes one participant and blocks future command actions.
func (r *RedisQueueRepository) KickParticipant(jamID string, actorUserID string, req model.ModerationCommandRequest) (model.SessionSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.getSessionState(jamID)
	if err != nil {
		return model.SessionSnapshot{}, err
	}
	if err := ensureActive(state); err != nil {
		return model.SessionSnapshot{}, err
	}
	if state.hostUserID != actorUserID {
		return model.SessionSnapshot{}, ErrHostOwnershipRequired
	}
	if req.TargetUserID == "" || req.TargetUserID == state.hostUserID {
		return model.SessionSnapshot{}, ErrModerationTargetInvalid
	}
	if _, exists := state.participants[req.TargetUserID]; !exists {
		return model.SessionSnapshot{}, ErrParticipantNotFound
	}

	delete(state.participants, req.TargetUserID)
	delete(state.mutedUsers, req.TargetUserID)
	state.kickedUsers[req.TargetUserID] = true
	state.sessionVersion++
	if err := r.saveDurableStateLocked(); err != nil {
		return model.SessionSnapshot{}, err
	}
	return buildSessionSnapshot(jamID, state), nil
}

// EndSession explicitly ends one session and requires host ownership.
func (r *RedisQueueRepository) EndSession(jamID string, actorUserID string) (model.SessionSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.getSessionState(jamID)
	if err != nil {
		return model.SessionSnapshot{}, err
	}
	if err := ensureActive(state); err != nil {
		return model.SessionSnapshot{}, err
	}
	if state.hostUserID != actorUserID {
		return model.SessionSnapshot{}, ErrHostOwnershipRequired
	}

	state.status = model.SessionStatusEnded
	state.endCause = "host_request"
	state.endedBy = actorUserID
	state.sessionVersion++
	if err := r.saveDurableStateLocked(); err != nil {
		return model.SessionSnapshot{}, err
	}
	return buildSessionSnapshot(jamID, state), nil
}

// SessionSnapshot returns current session metadata and participants.
func (r *RedisQueueRepository) SessionSnapshot(jamID string) (model.SessionSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.getSessionState(jamID)
	if err != nil {
		return model.SessionSnapshot{}, err
	}
	return buildSessionSnapshot(jamID, state), nil
}

// EnsureSessionActive validates that one session exists and is active.
func (r *RedisQueueRepository) EnsureSessionActive(jamID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.getSessionState(jamID)
	if err != nil {
		return err
	}
	return ensureActive(state)
}

// Add atomically appends one queue item, increments version, and records idempotency.
func (r *RedisQueueRepository) Add(jamID string, req model.AddQueueItemRequest) (model.QueueSnapshot, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if req.IdempotencyKey == "" {
		return model.QueueSnapshot{}, false, ErrIdempotencyKeyRequired
	}

	state, err := r.getSessionState(jamID)
	if err != nil {
		return model.QueueSnapshot{}, false, err
	}
	if err := ensureActive(state); err != nil {
		return model.QueueSnapshot{}, false, err
	}
	if err := ensureActorCanMutate(state, req.AddedBy); err != nil {
		return model.QueueSnapshot{}, false, err
	}

	if previous, ok := state.idempotencyResult[req.IdempotencyKey]; ok {
		return cloneSnapshot(previous.snapshot), true, nil
	}

	state.nextItemSequence++
	item := model.QueueItem{
		ItemID:  fmt.Sprintf("qi_%d", state.nextItemSequence),
		TrackID: req.TrackID,
		AddedBy: req.AddedBy,
	}

	state.items = append(state.items, item)
	state.queueVersion++

	snapshot := model.QueueSnapshot{
		JamID:        jamID,
		QueueVersion: state.queueVersion,
		Items:        cloneItems(state.items),
	}
	state.idempotencyResult[req.IdempotencyKey] = addResult{snapshot: snapshot}
	if err := r.saveDurableStateLocked(); err != nil {
		return model.QueueSnapshot{}, false, err
	}
	return cloneSnapshot(snapshot), false, nil
}

// Remove atomically deletes an item by ID and increments queue version by one.
func (r *RedisQueueRepository) Remove(jamID string, expectedVersion int64, itemID string, actorUserID string) (model.QueueSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.getSessionState(jamID)
	if err != nil {
		return model.QueueSnapshot{}, err
	}
	if err := ensureActive(state); err != nil {
		return model.QueueSnapshot{}, err
	}
	if err := ensureActorCanMutate(state, actorUserID); err != nil {
		return model.QueueSnapshot{}, err
	}
	if expectedVersion != state.queueVersion {
		return model.QueueSnapshot{}, &VersionConflictError{
			ExpectedQueueVersion: expectedVersion,
			CurrentQueueVersion:  state.queueVersion,
		}
	}
	index := slices.IndexFunc(state.items, func(item model.QueueItem) bool {
		return item.ItemID == itemID
	})
	if index == -1 {
		return model.QueueSnapshot{}, ErrQueueItemNotFound
	}

	state.items = append(state.items[:index], state.items[index+1:]...)
	state.queueVersion++
	if err := r.saveDurableStateLocked(); err != nil {
		return model.QueueSnapshot{}, err
	}

	return model.QueueSnapshot{
		JamID:        jamID,
		QueueVersion: state.queueVersion,
		Items:        cloneItems(state.items),
	}, nil
}

// Reorder atomically reorders queue items when expected version matches current.
func (r *RedisQueueRepository) Reorder(jamID string, expectedVersion int64, itemIDs []string, actorUserID string) (model.QueueSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.getSessionState(jamID)
	if err != nil {
		return model.QueueSnapshot{}, err
	}
	if err := ensureActive(state); err != nil {
		return model.QueueSnapshot{}, err
	}
	if err := ensureActorCanMutate(state, actorUserID); err != nil {
		return model.QueueSnapshot{}, err
	}
	if expectedVersion != state.queueVersion {
		return model.QueueSnapshot{}, &VersionConflictError{
			ExpectedQueueVersion: expectedVersion,
			CurrentQueueVersion:  state.queueVersion,
		}
	}

	if len(itemIDs) != len(state.items) {
		return model.QueueSnapshot{}, ErrQueueItemNotFound
	}

	byID := make(map[string]model.QueueItem, len(state.items))
	for _, item := range state.items {
		byID[item.ItemID] = item
	}

	reordered := make([]model.QueueItem, 0, len(itemIDs))
	seen := make(map[string]struct{}, len(itemIDs))
	for _, itemID := range itemIDs {
		item, ok := byID[itemID]
		if !ok {
			return model.QueueSnapshot{}, ErrQueueItemNotFound
		}
		if _, duplicate := seen[itemID]; duplicate {
			return model.QueueSnapshot{}, ErrQueueItemNotFound
		}
		seen[itemID] = struct{}{}
		reordered = append(reordered, item)
	}

	state.items = reordered
	state.queueVersion++
	if err := r.saveDurableStateLocked(); err != nil {
		return model.QueueSnapshot{}, err
	}

	return model.QueueSnapshot{
		JamID:        jamID,
		QueueVersion: state.queueVersion,
		Items:        cloneItems(state.items),
	}, nil
}

// Snapshot returns the latest committed queue state for a jam session.
func (r *RedisQueueRepository) Snapshot(jamID string) (model.QueueSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.getSessionState(jamID)
	if err != nil {
		return model.QueueSnapshot{}, err
	}
	return model.QueueSnapshot{
		JamID:        jamID,
		QueueVersion: state.queueVersion,
		Items:        cloneItems(state.items),
	}, nil
}

// ensureJamState lazily creates in-memory state for a jam session.
func (r *RedisQueueRepository) ensureJamState(jamID string) *jamQueueState {
	state, ok := r.jams[jamID]
	if ok {
		return state
	}

	state = &jamQueueState{
		items:             make([]model.QueueItem, 0),
		idempotencyResult: make(map[string]addResult),
		status:            model.SessionStatusActive,
		participants:      make(map[string]model.SessionRole),
		mutedUsers:        make(map[string]bool),
		kickedUsers:       make(map[string]bool),
	}
	r.jams[jamID] = state
	return state
}

func ensureActive(state *jamQueueState) error {
	if state.status == model.SessionStatusEnded {
		return ErrSessionEnded
	}
	return nil
}

func (r *RedisQueueRepository) getSessionState(jamID string) (*jamQueueState, error) {
	state, ok := r.jams[jamID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return state, nil
}

func buildSessionSnapshot(jamID string, state *jamQueueState) model.SessionSnapshot {
	participants := make([]model.SessionParticipant, 0, len(state.participants))
	for userID, role := range state.participants {
		participants = append(participants, model.SessionParticipant{
			UserID: userID,
			Role:   role,
			Muted:  state.mutedUsers[userID],
		})
	}
	sort.Slice(participants, func(i, j int) bool {
		return participants[i].UserID < participants[j].UserID
	})

	return model.SessionSnapshot{
		JamID:          jamID,
		Status:         state.status,
		HostUserID:     state.hostUserID,
		Participants:   participants,
		SessionVersion: state.sessionVersion,
		EndCause:       state.endCause,
		EndedBy:        state.endedBy,
	}
}

func ensureActorCanMutate(state *jamQueueState, actorUserID string) error {
	if actorUserID == "" {
		return ErrModerationBlocked
	}
	if state.kickedUsers[actorUserID] {
		return ErrModerationBlocked
	}
	if _, exists := state.participants[actorUserID]; !exists {
		return ErrModerationBlocked
	}
	if state.mutedUsers[actorUserID] {
		return ErrModerationBlocked
	}
	return nil
}

// cloneItems returns a copy to keep repository internal state immutable externally.
func cloneItems(items []model.QueueItem) []model.QueueItem {
	copied := make([]model.QueueItem, len(items))
	copy(copied, items)
	return copied
}

// cloneSnapshot returns a deep copy of the snapshot.
func cloneSnapshot(snapshot model.QueueSnapshot) model.QueueSnapshot {
	snapshot.Items = cloneItems(snapshot.Items)
	return snapshot
}
