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

func TestPublishStateTransition_UsesPlaybackTopicAndSessionKey(t *testing.T) {
	pub := &capturePublisher{}
	producer := NewProducer(pub)

	err := producer.PublishStateTransition(context.Background(), "jam_1", 3, "jam.playback.changed", map[string]string{"state": "pause"})
	if err != nil {
		t.Fatalf("PublishStateTransition() error = %v", err)
	}
	if pub.topic != sharedkafka.TopicJamPlayback {
		t.Fatalf("unexpected topic: %s", pub.topic)
	}
	if pub.key != "jam_1" {
		t.Fatalf("unexpected key: %s", pub.key)
	}
}

func TestPublishStateTransition_RejectsMissingSessionID(t *testing.T) {
	pub := &capturePublisher{}
	producer := NewProducer(pub)

	if err := producer.PublishStateTransition(context.Background(), "", 3, "jam.playback.changed", map[string]string{"state": "pause"}); err == nil {
		t.Fatal("expected validation error")
	}
}
