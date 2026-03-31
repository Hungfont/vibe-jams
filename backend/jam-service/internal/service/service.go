package service

import (
	"context"
	"errors"
	"fmt"

	"video-streaming/backend/jams/internal/model"
	"video-streaming/backend/jams/internal/repository"
)

var (
	// ErrInvalidRequest indicates payload validation failure.
	ErrInvalidRequest = errors.New("invalid request")
	// ErrSessionEnded indicates write attempted on ended session.
	ErrSessionEnded = errors.New("session ended")
)

// QueueRepository abstracts queue persistence for service business logic.
type QueueRepository interface {
	CreateSession(hostUserID string) (model.SessionSnapshot, error)
	JoinSession(jamID string, userID string) (model.SessionSnapshot, error)
	LeaveSession(jamID string, userID string) (model.SessionSnapshot, error)
	EndSession(jamID string, actorUserID string) (model.SessionSnapshot, error)
	SessionSnapshot(jamID string) (model.SessionSnapshot, error)
	EnsureSessionActive(jamID string) error

	Add(jamID string, req model.AddQueueItemRequest) (model.QueueSnapshot, bool, error)
	Remove(jamID string, itemID string) (model.QueueSnapshot, error)
	Reorder(jamID string, expectedVersion int64, itemIDs []string) (model.QueueSnapshot, error)
	Snapshot(jamID string) (model.QueueSnapshot, error)
}

// EventProducer abstracts queue/session event publishing.
type EventProducer interface {
	PublishQueueEvent(ctx context.Context, sessionID string, aggregateVersion int64, eventType string, payload any) error
	PublishSessionEvent(ctx context.Context, sessionID string, aggregateVersion int64, eventType string, payload any) error
}

// Service orchestrates queue command validation and mutation behavior.
type Service struct {
	repo   QueueRepository
	events EventProducer
}

// New builds a queue service instance with injected repository dependency.
func New(repo QueueRepository, events EventProducer) *Service {
	return &Service{repo: repo, events: events}
}

// CreateSession starts a new jam session owned by host user.
func (s *Service) CreateSession(hostUserID string) (model.SessionSnapshot, error) {
	if hostUserID == "" {
		return model.SessionSnapshot{}, fmt.Errorf("%w: host user is required", ErrInvalidRequest)
	}

	session, err := s.repo.CreateSession(hostUserID)
	if err != nil {
		return model.SessionSnapshot{}, err
	}
	if s.events != nil {
		if err := s.events.PublishSessionEvent(context.Background(), session.JamID, session.SessionVersion, "jam.session.created", map[string]any{
			"hostUserId": hostUserID,
			"status":     session.Status,
		}); err != nil {
			return model.SessionSnapshot{}, err
		}
	}
	return session, nil
}

// JoinSession adds a user into active session membership.
func (s *Service) JoinSession(jamID string, userID string) (model.SessionSnapshot, error) {
	if jamID == "" || userID == "" {
		return model.SessionSnapshot{}, fmt.Errorf("%w: jamId and userId are required", ErrInvalidRequest)
	}

	session, err := s.repo.JoinSession(jamID, userID)
	if err != nil {
		return model.SessionSnapshot{}, err
	}
	if s.events != nil {
		if err := s.events.PublishSessionEvent(context.Background(), jamID, session.SessionVersion, "jam.session.joined", map[string]any{
			"userId": userID,
		}); err != nil {
			return model.SessionSnapshot{}, err
		}
	}
	return session, nil
}

// LeaveSession removes a user from active membership.
// Host leave transitions the session to ended state.
func (s *Service) LeaveSession(jamID string, userID string) (model.SessionSnapshot, error) {
	if jamID == "" || userID == "" {
		return model.SessionSnapshot{}, fmt.Errorf("%w: jamId and userId are required", ErrInvalidRequest)
	}

	session, err := s.repo.LeaveSession(jamID, userID)
	if err != nil {
		return model.SessionSnapshot{}, err
	}
	if s.events != nil {
		eventType := "jam.session.left"
		if session.Status == model.SessionStatusEnded {
			eventType = "jam.session.ended"
		}
		if err := s.events.PublishSessionEvent(context.Background(), jamID, session.SessionVersion, eventType, map[string]any{
			"userId":   userID,
			"status":   session.Status,
			"endCause": session.EndCause,
		}); err != nil {
			return model.SessionSnapshot{}, err
		}
	}
	return session, nil
}

// EndSession explicitly ends an active session. Only host may end.
func (s *Service) EndSession(jamID string, actorUserID string) (model.SessionSnapshot, error) {
	if jamID == "" || actorUserID == "" {
		return model.SessionSnapshot{}, fmt.Errorf("%w: jamId and actor user are required", ErrInvalidRequest)
	}

	session, err := s.repo.EndSession(jamID, actorUserID)
	if err != nil {
		return model.SessionSnapshot{}, err
	}
	if s.events != nil {
		if err := s.events.PublishSessionEvent(context.Background(), jamID, session.SessionVersion, "jam.session.ended", map[string]any{
			"endedBy":  actorUserID,
			"endCause": session.EndCause,
		}); err != nil {
			return model.SessionSnapshot{}, err
		}
	}
	return session, nil
}

// Add validates and appends a queue item with idempotency protection.
func (s *Service) Add(jamID string, req model.AddQueueItemRequest) (model.QueueSnapshot, bool, error) {
	if jamID == "" || req.TrackID == "" || req.AddedBy == "" {
		return model.QueueSnapshot{}, false, fmt.Errorf("%w: jamId, trackId and addedBy are required", ErrInvalidRequest)
	}
	if req.IdempotencyKey == "" {
		return model.QueueSnapshot{}, false, fmt.Errorf("%w: idempotencyKey is required", ErrInvalidRequest)
	}
	if err := s.repo.EnsureSessionActive(jamID); err != nil {
		return model.QueueSnapshot{}, false, err
	}

	snapshot, fromCache, err := s.repo.Add(jamID, req)
	if err != nil {
		return model.QueueSnapshot{}, false, err
	}
	if err := s.publishMutationEvents(jamID, snapshot.QueueVersion, "jam.queue.item.added", map[string]any{
		"trackId": req.TrackID,
		"addedBy": req.AddedBy,
	}); err != nil {
		return model.QueueSnapshot{}, false, err
	}

	return snapshot, fromCache, nil
}

// Remove validates and removes an existing queue item.
func (s *Service) Remove(jamID string, req model.RemoveQueueItemRequest) (model.QueueSnapshot, error) {
	if jamID == "" || req.ItemID == "" {
		return model.QueueSnapshot{}, fmt.Errorf("%w: jamId and itemId are required", ErrInvalidRequest)
	}
	if err := s.repo.EnsureSessionActive(jamID); err != nil {
		return model.QueueSnapshot{}, err
	}

	snapshot, err := s.repo.Remove(jamID, req.ItemID)
	if err != nil {
		return model.QueueSnapshot{}, err
	}
	if err := s.publishMutationEvents(jamID, snapshot.QueueVersion, "jam.queue.item.removed", map[string]any{
		"itemId": req.ItemID,
	}); err != nil {
		return model.QueueSnapshot{}, err
	}

	return snapshot, nil
}

// Reorder validates request and applies optimistic concurrency semantics.
func (s *Service) Reorder(jamID string, req model.ReorderQueueRequest) (model.QueueSnapshot, error) {
	if jamID == "" {
		return model.QueueSnapshot{}, fmt.Errorf("%w: jamId is required", ErrInvalidRequest)
	}
	if len(req.ItemIDs) == 0 {
		return model.QueueSnapshot{}, fmt.Errorf("%w: itemIds is required", ErrInvalidRequest)
	}
	if err := s.repo.EnsureSessionActive(jamID); err != nil {
		return model.QueueSnapshot{}, err
	}

	snapshot, err := s.repo.Reorder(jamID, req.ExpectedQueueVersion, req.ItemIDs)
	if err != nil {
		return model.QueueSnapshot{}, err
	}
	if err := s.publishMutationEvents(jamID, snapshot.QueueVersion, "jam.queue.reordered", map[string]any{
		"itemIds": req.ItemIDs,
	}); err != nil {
		return model.QueueSnapshot{}, err
	}

	return snapshot, nil
}

// Snapshot validates jam ID and returns latest committed queue state.
func (s *Service) Snapshot(jamID string) (model.QueueSnapshot, error) {
	if jamID == "" {
		return model.QueueSnapshot{}, fmt.Errorf("%w: jamId is required", ErrInvalidRequest)
	}

	return s.repo.Snapshot(jamID)
}

// IsVersionConflict reports whether an error is version conflict.
func IsVersionConflict(err error) bool {
	return errors.Is(err, repository.ErrVersionConflict)
}

// IsNotFound reports whether an error is queue-item-not-found.
func IsNotFound(err error) bool {
	return errors.Is(err, repository.ErrQueueItemNotFound) ||
		errors.Is(err, repository.ErrSessionNotFound) ||
		errors.Is(err, repository.ErrParticipantNotFound)
}

// IsHostOnly reports whether operation requires host ownership.
func IsHostOnly(err error) bool {
	return errors.Is(err, repository.ErrHostOwnershipRequired)
}

// IsSessionEnded reports whether operation was blocked by ended state.
func IsSessionEnded(err error) bool {
	return errors.Is(err, ErrSessionEnded) || errors.Is(err, repository.ErrSessionEnded)
}

// IsInvalidRequest reports whether an error is a validation failure.
func IsInvalidRequest(err error) bool {
	return errors.Is(err, ErrInvalidRequest) || errors.Is(err, repository.ErrIdempotencyKeyRequired)
}

func (s *Service) publishMutationEvents(jamID string, queueVersion int64, queueEventType string, payload any) error {
	if s.events == nil {
		return nil
	}

	ctx := context.Background()
	if err := s.events.PublishQueueEvent(ctx, jamID, queueVersion, queueEventType, payload); err != nil {
		return err
	}

	return s.events.PublishSessionEvent(ctx, jamID, queueVersion, "jam.session.updated", map[string]any{
		"queueVersion": queueVersion,
	})
}
