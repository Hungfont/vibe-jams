package server

import (
	"context"

	sharedevent "video-streaming/backend/shared/event"
)

// ModerationEventHook is an extension point for moderation-abuse heuristics.
type ModerationEventHook interface {
	HandleModerationEvent(ctx context.Context, envelope sharedevent.Envelope) error
}

// NoopModerationEventHook is the default hook implementation.
type NoopModerationEventHook struct{}

// HandleModerationEvent intentionally does nothing.
func (NoopModerationEventHook) HandleModerationEvent(_ context.Context, _ sharedevent.Envelope) error {
	return nil
}
