package service

import (
	"testing"

	"video-streaming/backend/jams/internal/kafka"
	"video-streaming/backend/jams/internal/repository"
	sharedevent "video-streaming/backend/shared/event"
	sharedkafka "video-streaming/backend/shared/kafka"
)

func TestLifecycleEvents_ArePublishedToSessionTopic(t *testing.T) {
	t.Parallel()

	repo := repository.NewRedisQueueRepository()
	pub := &kafka.InMemoryPublisher{}
	svc := New(repo, kafka.NewProducer(pub))

	created, err := svc.CreateSession("host_1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := svc.JoinSession(created.JamID, "member_1"); err != nil {
		t.Fatalf("join session: %v", err)
	}
	if _, err := svc.LeaveSession(created.JamID, "member_1"); err != nil {
		t.Fatalf("leave session: %v", err)
	}
	if _, err := svc.EndSession(created.JamID, "host_1"); err != nil {
		t.Fatalf("end session: %v", err)
	}

	if len(pub.Records) < 4 {
		t.Fatalf("expected at least 4 records, got %d", len(pub.Records))
	}
	sessionTopicCount := 0
	for _, record := range pub.Records {
		if record.Topic != sharedkafka.TopicJamSession {
			continue
		}
		sessionTopicCount++
		if _, err := sharedevent.UnmarshalEnvelope(record.Value); err != nil {
			t.Fatalf("unmarshal envelope: %v", err)
		}
	}
	if sessionTopicCount < 4 {
		t.Fatalf("expected at least 4 session-topic records, got %d", sessionTopicCount)
	}
}
