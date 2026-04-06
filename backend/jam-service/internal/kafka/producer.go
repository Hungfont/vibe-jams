package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"video-streaming/backend/shared/event"
	"video-streaming/backend/shared/kafka"
)

var ErrSessionIDKeyRequired = errors.New("sessionId key is required")

// Publisher writes serialized messages to Kafka.
type Publisher interface {
	Publish(ctx context.Context, topic string, key string, value []byte) error
}

// Producer publishes session and queue events for jam-service.
type Producer struct {
	pub Publisher
}

// NewProducer builds a producer adapter from low-level publisher.
func NewProducer(pub Publisher) *Producer {
	return &Producer{pub: pub}
}

// PublishQueueEvent emits queue mutation events keyed by session ID.
func (p *Producer) PublishQueueEvent(ctx context.Context, sessionID string, aggregateVersion int64, eventType string, payload any) error {
	if sessionID == "" {
		return ErrSessionIDKeyRequired
	}

	env := event.Envelope{
		EventID:          fmt.Sprintf("queue-%d", time.Now().UnixNano()),
		EventType:        eventType,
		SessionID:        sessionID,
		AggregateVersion: aggregateVersion,
		OccurredAt:       time.Now().UTC(),
		Payload:          event.MustPayload(payload),
	}
	raw, err := event.MarshalEnvelope(env, true)
	if err != nil {
		return err
	}

	return p.pub.Publish(ctx, kafka.TopicJamQueue, sessionID, raw)
}

// PublishSessionEvent emits session lifecycle events keyed by session ID.
func (p *Producer) PublishSessionEvent(ctx context.Context, sessionID string, aggregateVersion int64, eventType string, payload any) error {
	if sessionID == "" {
		return ErrSessionIDKeyRequired
	}

	env := event.Envelope{
		EventID:          fmt.Sprintf("session-%d", time.Now().UnixNano()),
		EventType:        eventType,
		SessionID:        sessionID,
		AggregateVersion: aggregateVersion,
		OccurredAt:       time.Now().UTC(),
		Payload:          event.MustPayload(payload),
	}
	raw, err := event.MarshalEnvelope(env, true)
	if err != nil {
		return err
	}

	return p.pub.Publish(ctx, kafka.TopicJamSession, sessionID, raw)
}

// PublishModerationEvent emits moderation audit events keyed by session ID.
func (p *Producer) PublishModerationEvent(ctx context.Context, sessionID string, aggregateVersion int64, eventType string, payload any) error {
	if sessionID == "" {
		return ErrSessionIDKeyRequired
	}

	env := event.Envelope{
		EventID:          fmt.Sprintf("moderation-%d", time.Now().UnixNano()),
		EventType:        eventType,
		SessionID:        sessionID,
		AggregateVersion: aggregateVersion,
		OccurredAt:       time.Now().UTC(),
		Payload:          event.MustPayload(payload),
	}
	raw, err := event.MarshalEnvelope(env, true)
	if err != nil {
		return err
	}

	return p.pub.Publish(ctx, kafka.TopicJamModeration, sessionID, raw)
}
