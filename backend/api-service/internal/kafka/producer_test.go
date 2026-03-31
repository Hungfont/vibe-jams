package kafka

import (
	"context"
	"testing"

	sharedkafka "video-streaming/backend/shared/kafka"
)

type capturePublisher struct {
	topic string
	key   string
	value []byte
}

func (c *capturePublisher) Publish(_ context.Context, topic string, key string, value []byte) error {
	c.topic = topic
	c.key = key
	c.value = value
	return nil
}

func TestPublishUserAction_UsesAnalyticsTopicAndUserKey(t *testing.T) {
	pub := &capturePublisher{}
	producer := NewProducer(pub)

	err := producer.PublishUserAction(context.Background(), "user_1", "analytics.user.action", map[string]string{"action": "queue_add"})
	if err != nil {
		t.Fatalf("PublishUserAction() error = %v", err)
	}
	if pub.topic != sharedkafka.TopicAnalyticsUser {
		t.Fatalf("unexpected topic: %s", pub.topic)
	}
	if pub.key != "user_1" {
		t.Fatalf("unexpected key: %s", pub.key)
	}
}

func TestPublishUserAction_RejectsMissingUserID(t *testing.T) {
	pub := &capturePublisher{}
	producer := NewProducer(pub)

	if err := producer.PublishUserAction(context.Background(), "", "analytics.user.action", map[string]string{}); err == nil {
		t.Fatal("expected validation error")
	}
}
