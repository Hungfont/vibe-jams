package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"video-streaming/backend/playback-service/internal/model"
	"video-streaming/backend/playback-service/internal/repository"
	sharedcatalog "video-streaming/backend/shared/catalog"
)

var (
	// ErrInvalidRequest indicates payload validation failure.
	ErrInvalidRequest = errors.New("invalid request")
	// ErrHostOnly indicates command requires host role.
	ErrHostOnly = errors.New("host only")
	// ErrSessionEnded indicates command is blocked by ended session.
	ErrSessionEnded = errors.New("session ended")
	// ErrVersionConflict indicates expected queue version is stale.
	ErrVersionConflict = errors.New("version conflict")
	// ErrTrackNotFound indicates the requested track does not exist.
	ErrTrackNotFound = errors.New("track not found")
	// ErrTrackUnavailable indicates the requested track cannot be played.
	ErrTrackUnavailable = errors.New("track unavailable")
)

// StateRepository abstracts queue/playback metadata persistence.
type StateRepository interface {
	SessionStatus(sessionID string) (string, error)
	HostUserID(sessionID string) (string, error)
	QueueVersion(sessionID string) (int64, error)
	PlaybackEpoch(sessionID string) (int64, error)
	ApplyCommand(sessionID string, command string, positionMS int64, actorUserID string, clientEventID string) (model.PlaybackTransition, error)
}

// ConflictRetryGuidance contains authoritative metadata for stale command recovery.
type ConflictRetryGuidance struct {
	CurrentQueueVersion int64
	PlaybackEpoch       int64
}

type versionConflictError struct {
	guidance ConflictRetryGuidance
}

func (e *versionConflictError) Error() string {
	return "stale queue version"
}

func (e *versionConflictError) Is(target error) bool {
	return target == ErrVersionConflict
}

// ConflictRetryFromError extracts retry guidance when a version conflict occurs.
func ConflictRetryFromError(err error) (ConflictRetryGuidance, bool) {
	var conflictErr *versionConflictError
	if !errors.As(err, &conflictErr) {
		return ConflictRetryGuidance{}, false
	}

	return conflictErr.guidance, true
}

// EventProducer publishes playback transition events.
type EventProducer interface {
	PublishStateTransition(ctx context.Context, sessionID string, aggregateVersion int64, eventType string, payload any) error
}

// Service orchestrates command validation, execution, and event emission.
type Service struct {
	repo                     StateRepository
	events                   EventProducer
	catalogValidator         sharedcatalog.Validator
	catalogValidationEnabled bool
}

// New builds a playback command service with injected dependencies.
func New(repo StateRepository, events EventProducer) *Service {
	return NewWithCatalogValidator(repo, events, nil, false)
}

// NewWithCatalogValidator builds a playback service with optional catalog pre-checks.
func NewWithCatalogValidator(repo StateRepository, events EventProducer, catalogValidator sharedcatalog.Validator, enabled bool) *Service {
	return &Service{
		repo:                     repo,
		events:                   events,
		catalogValidator:         catalogValidator,
		catalogValidationEnabled: enabled,
	}
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
	if err := s.validateTrack(ctx, req.TrackID); err != nil {
		return model.CommandAcceptedResponse{}, err
	}
	status, err := s.repo.SessionStatus(sessionID)
	if err != nil {
		return model.CommandAcceptedResponse{}, err
	}
	if strings.EqualFold(status, "ended") {
		return model.CommandAcceptedResponse{}, ErrSessionEnded
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
	playbackEpoch, err := s.repo.PlaybackEpoch(sessionID)
	if err != nil {
		return model.CommandAcceptedResponse{}, err
	}
	if req.ExpectedQueueVersion != queueVersion {
		return model.CommandAcceptedResponse{}, &versionConflictError{
			guidance: ConflictRetryGuidance{
				CurrentQueueVersion: queueVersion,
				PlaybackEpoch:       playbackEpoch,
			},
		}
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
			"playbackEpoch":    transition.PlaybackEpoch,
			"actorUserId":      transition.ActorUserID,
			"clientEventId":    transition.ClientEventID,
			"aggregateVersion": transition.AggregateVersion,
		}); err != nil {
			return model.CommandAcceptedResponse{}, err
		}
	}

	return model.CommandAcceptedResponse{
		Accepted:      true,
		QueueVersion:  transition.QueueVersion,
		PlaybackEpoch: transition.PlaybackEpoch,
	}, nil
}

func (s *Service) validateTrack(ctx context.Context, trackID string) error {
	if !s.catalogValidationEnabled || s.catalogValidator == nil {
		return nil
	}
	if strings.TrimSpace(trackID) == "" {
		return nil
	}

	_, err := s.catalogValidator.ValidateTrack(ctx, trackID)
	if err == nil {
		return nil
	}

	if errors.Is(err, sharedcatalog.ErrTrackNotFound) {
		return ErrTrackNotFound
	}
	if errors.Is(err, sharedcatalog.ErrTrackUnavailable) {
		return ErrTrackUnavailable
	}

	return err
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

// IsTrackNotFound reports whether command failed due to unknown track.
func IsTrackNotFound(err error) bool {
	return errors.Is(err, ErrTrackNotFound)
}

// IsTrackUnavailable reports whether command failed due to unavailable track.
func IsTrackUnavailable(err error) bool {
	return errors.Is(err, ErrTrackUnavailable)
}

// IsSessionEnded reports whether command is blocked by ended state.
func IsSessionEnded(err error) bool {
	return errors.Is(err, ErrSessionEnded) || errors.Is(err, repository.ErrSessionEnded)
}

func normalizeCommand(command string) string {
	return strings.ToLower(strings.TrimSpace(command))
}
