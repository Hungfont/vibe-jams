package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"video-streaming/backend/playback-service/internal/model"
	"video-streaming/backend/playback-service/internal/repository"
)

var (
	// ErrInvalidRequest indicates payload validation failure.
	ErrInvalidRequest = errors.New("invalid request")
	// ErrHostOnly indicates command requires host role.
	ErrHostOnly = errors.New("host only")
	// ErrVersionConflict indicates expected queue version is stale.
	ErrVersionConflict = errors.New("version conflict")
)

// StateRepository abstracts queue/playback metadata persistence.
type StateRepository interface {
	HostUserID(sessionID string) (string, error)
	QueueVersion(sessionID string) (int64, error)
	ApplyCommand(sessionID string, command string, positionMS int64, actorUserID string, clientEventID string) (model.PlaybackTransition, error)
}

// EventProducer publishes playback transition events.
type EventProducer interface {
	PublishStateTransition(ctx context.Context, sessionID string, aggregateVersion int64, eventType string, payload any) error
}

// Service orchestrates command validation, execution, and event emission.
type Service struct {
	repo   StateRepository
	events EventProducer
}

// New builds a playback command service with injected dependencies.
func New(repo StateRepository, events EventProducer) *Service {
	return &Service{repo: repo, events: events}
}

// ExecuteCommand validates request, enforces host policy, and emits transition event.
func (s *Service) ExecuteCommand(ctx context.Context, sessionID string, actorUserID string, req model.PlaybackCommandRequest) (model.CommandAcceptedResponse, error) {
	if strings.TrimSpace(sessionID) == "" {
		return model.CommandAcceptedResponse{}, fmt.Errorf("%w: sessionId is required", ErrInvalidRequest)
	}
	if strings.TrimSpace(actorUserID) == "" {
		return model.CommandAcceptedResponse{}, fmt.Errorf("%w: actor user is required", ErrInvalidRequest)
	}
	if err := req.Validate(); err != nil {
		return model.CommandAcceptedResponse{}, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	hostUserID, err := s.repo.HostUserID(sessionID)
	if err != nil {
		return model.CommandAcceptedResponse{}, err
	}
	if hostUserID != actorUserID {
		return model.CommandAcceptedResponse{}, ErrHostOnly
	}

	queueVersion, err := s.repo.QueueVersion(sessionID)
	if err != nil {
		return model.CommandAcceptedResponse{}, err
	}
	if req.ExpectedQueueVersion != queueVersion {
		return model.CommandAcceptedResponse{}, ErrVersionConflict
	}

	transition, err := s.repo.ApplyCommand(sessionID, normalizeCommand(req.Command), req.PositionMS, actorUserID, req.ClientEventID)
	if err != nil {
		return model.CommandAcceptedResponse{}, err
	}

	if s.events != nil {
		if err := s.events.PublishStateTransition(ctx, sessionID, transition.AggregateVersion, "jam.playback.updated", map[string]any{
			"command":          transition.Command,
			"state":            transition.State,
			"positionMs":       transition.PositionMS,
			"queueVersion":     transition.QueueVersion,
			"actorUserId":      transition.ActorUserID,
			"clientEventId":    transition.ClientEventID,
			"aggregateVersion": transition.AggregateVersion,
		}); err != nil {
			return model.CommandAcceptedResponse{}, err
		}
	}

	return model.CommandAcceptedResponse{Accepted: true}, nil
}

// IsInvalidRequest reports whether validation failed.
func IsInvalidRequest(err error) bool {
	return errors.Is(err, ErrInvalidRequest) || errors.Is(err, model.ErrInvalidCommand)
}

// IsHostOnly reports whether request failed host-only policy.
func IsHostOnly(err error) bool {
	return errors.Is(err, ErrHostOnly)
}

// IsVersionConflict reports whether staleness check failed.
func IsVersionConflict(err error) bool {
	return errors.Is(err, ErrVersionConflict)
}

// IsNotFound reports whether session state is missing.
func IsNotFound(err error) bool {
	return errors.Is(err, repository.ErrSessionNotFound)
}

func normalizeCommand(command string) string {
	return strings.ToLower(strings.TrimSpace(command))
}
