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

// Producer publishes playback transition events.
type Producer struct {
	pub Publisher
}

// NewProducer creates a playback Kafka producer adapter.
func NewProducer(pub Publisher) *Producer {
	return &Producer{pub: pub}
}

// PublishStateTransition validates and publishes a playback event.
func (p *Producer) PublishStateTransition(ctx context.Context, sessionID string, aggregateVersion int64, eventType string, payload any) error {
	if sessionID == "" {
		return ErrSessionIDKeyRequired
	}

	env := event.Envelope{
		EventID:          fmt.Sprintf("playback-%d", time.Now().UnixNano()),
		EventType:        eventType,
		SessionID:        sessionID,
		AggregateVersion: aggregateVersion,
		OccurredAt:       time.Now().UTC(),
		Payload:          event.MustPayload(payload),
	}

	encoded, err := event.MarshalEnvelope(env, true)
	if err != nil {
		return err
	}

	return p.pub.Publish(ctx, kafka.TopicJamPlayback, sessionID, encoded)
}
