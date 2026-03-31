package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"video-streaming/backend/shared/event"
	"video-streaming/backend/shared/kafka"
)

var ErrUserIDKeyRequired = errors.New("userId key is required")

// Publisher writes serialized messages to Kafka.
type Publisher interface {
	Publish(ctx context.Context, topic string, key string, value []byte) error
}

// Producer publishes analytics action events.
type Producer struct {
	pub Publisher
}

// NewProducer creates an analytics Kafka producer adapter.
func NewProducer(pub Publisher) *Producer {
	return &Producer{pub: pub}
}

// PublishUserAction validates and publishes analytics user action event.
func (p *Producer) PublishUserAction(ctx context.Context, userID string, eventType string, payload any) error {
	if userID == "" {
		return ErrUserIDKeyRequired
	}

	env := event.Envelope{
		EventID:     fmt.Sprintf("analytics-%d", time.Now().UnixNano()),
		EventType:   eventType,
		ActorUserID: userID,
		OccurredAt:  time.Now().UTC(),
		Payload:     event.MustPayload(payload),
	}

	encoded, err := event.MarshalEnvelope(env, false)
	if err != nil {
		return err
	}

	return p.pub.Publish(ctx, kafka.TopicAnalyticsUser, userID, encoded)
}
