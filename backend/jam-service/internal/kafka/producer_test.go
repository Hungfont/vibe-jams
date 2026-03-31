package kafka

import (
	"context"
	"testing"

	sharedkafka "video-streaming/backend/shared/kafka"
)

func TestPublishQueueEvent_UsesSessionKeyAndQueueTopic(t *testing.T) {
	pub := &InMemoryPublisher{}
	producer := NewProducer(pub)

	err := producer.PublishQueueEvent(context.Background(), "jam_1", 2, "jam.queue.item.added", map[string]string{"trackId": "t1"})
	if err != nil {
		t.Fatalf("PublishQueueEvent() error = %v", err)
	}
	if len(pub.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(pub.Records))
	}
	if pub.Records[0].Topic != sharedkafka.TopicJamQueue {
		t.Fatalf("unexpected topic: %s", pub.Records[0].Topic)
	}
	if pub.Records[0].Key != "jam_1" {
		t.Fatalf("unexpected key: %s", pub.Records[0].Key)
	}
}

func TestPublishSessionEvent_MissingSessionID(t *testing.T) {
	pub := &InMemoryPublisher{}
	producer := NewProducer(pub)

	if err := producer.PublishSessionEvent(context.Background(), "", 1, "jam.session.updated", map[string]string{}); err == nil {
		t.Fatal("expected validation error")
	}
}
