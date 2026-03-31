package repository

import (
	"errors"
	"fmt"
	"slices"
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
)

type addResult struct {
	snapshot model.QueueSnapshot
}

type jamQueueState struct {
	items             []model.QueueItem
	queueVersion      int64
	nextItemSequence  int64
	idempotencyResult map[string]addResult
}

// RedisQueueRepository stores queue state using in-memory maps with Redis-like key model.
type RedisQueueRepository struct {
	mu   sync.Mutex
	jams map[string]*jamQueueState
}

// NewRedisQueueRepository builds a repository implementation used by queue service.
func NewRedisQueueRepository() *RedisQueueRepository {
	return &RedisQueueRepository{
		jams: make(map[string]*jamQueueState),
	}
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

// Add atomically appends one queue item, increments version, and records idempotency.
func (r *RedisQueueRepository) Add(jamID string, req model.AddQueueItemRequest) (model.QueueSnapshot, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if req.IdempotencyKey == "" {
		return model.QueueSnapshot{}, false, ErrIdempotencyKeyRequired
	}

	state := r.ensureJamState(jamID)

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
	return cloneSnapshot(snapshot), false, nil
}

// Remove atomically deletes an item by ID and increments queue version by one.
func (r *RedisQueueRepository) Remove(jamID string, itemID string) (model.QueueSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state := r.ensureJamState(jamID)
	index := slices.IndexFunc(state.items, func(item model.QueueItem) bool {
		return item.ItemID == itemID
	})
	if index == -1 {
		return model.QueueSnapshot{}, ErrQueueItemNotFound
	}

	state.items = append(state.items[:index], state.items[index+1:]...)
	state.queueVersion++

	return model.QueueSnapshot{
		JamID:        jamID,
		QueueVersion: state.queueVersion,
		Items:        cloneItems(state.items),
	}, nil
}

// Reorder atomically reorders queue items when expected version matches current.
func (r *RedisQueueRepository) Reorder(jamID string, expectedVersion int64, itemIDs []string) (model.QueueSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state := r.ensureJamState(jamID)
	if expectedVersion != state.queueVersion {
		return model.QueueSnapshot{}, ErrVersionConflict
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

	state := r.ensureJamState(jamID)
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
	}
	r.jams[jamID] = state
	return state
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
