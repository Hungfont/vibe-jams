package service

import (
	"context"
	"encoding/json"
	"testing"

	"video-streaming/backend/jams/internal/kafka"
	"video-streaming/backend/jams/internal/model"
	"video-streaming/backend/jams/internal/repository"
	sharedauth "video-streaming/backend/shared/auth"
	sharedevent "video-streaming/backend/shared/event"
	sharedkafka "video-streaming/backend/shared/kafka"
)

func TestPermissionProjectionDefaultsAndEventPublication(t *testing.T) {
	t.Parallel()

	repo := repository.NewRedisQueueRepository()
	pub := &kafka.InMemoryPublisher{}
	svc := New(repo, kafka.NewProducer(pub))

	session, err := svc.CreateSession("host_1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if session.Permissions.CanControlPlayback || session.Permissions.CanReorderQueue || session.Permissions.CanChangeVolume {
		t.Fatalf("expected default false permissions, got %+v", session.Permissions)
	}

	snapshot, err := svc.UpdatePermissions(context.Background(), session.JamID, sharedauth.Claims{
		UserID:       "host_1",
		Plan:         "premium",
		SessionState: sharedauth.SessionStateValid,
	}, model.PermissionUpdateRequest{CanControlPlayback: boolPtr(true)})
	if err != nil {
		t.Fatalf("update permissions: %v", err)
	}
	if !snapshot.Permissions.CanControlPlayback {
		t.Fatal("expected canControlPlayback=true after update")
	}

	foundPermissionEvent := false
	for _, record := range pub.Records {
		if record.Topic != sharedkafka.TopicJamPermission {
			continue
		}
		env, err := sharedevent.UnmarshalEnvelope(record.Value)
		if err != nil {
			t.Fatalf("unmarshal permission envelope: %v", err)
		}
		if env.EventType != "jam.permission.updated" {
			continue
		}
		var payload struct {
			CanControlPlayback bool `json:"canControlPlayback"`
		}
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			t.Fatalf("unmarshal permission payload: %v", err)
		}
		if !payload.CanControlPlayback {
			t.Fatal("permission event payload mismatch")
		}
		foundPermissionEvent = true
	}
	if !foundPermissionEvent {
		t.Fatal("expected jam.permission.updated event")
	}
}

func TestGuestReorderDeniedUntilPermissionEnabled(t *testing.T) {
	t.Parallel()

	repo := repository.NewRedisQueueRepository()
	svc := New(repo, nil)

	session, err := svc.CreateSession("host_1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := svc.JoinSession(session.JamID, "member_1"); err != nil {
		t.Fatalf("join session: %v", err)
	}

	_, _, err = svc.Add(session.JamID, model.AddQueueItemRequest{TrackID: "trk_1", AddedBy: "host_1", IdempotencyKey: "k1"})
	if err != nil {
		t.Fatalf("add queue item 1: %v", err)
	}
	second, _, err := svc.Add(session.JamID, model.AddQueueItemRequest{TrackID: "trk_2", AddedBy: "host_1", IdempotencyKey: "k2"})
	if err != nil {
		t.Fatalf("add queue item 2: %v", err)
	}

	_, err = svc.Reorder(session.JamID, model.ReorderQueueRequest{
		ItemIDs:              []string{second.Items[1].ItemID, second.Items[0].ItemID},
		ExpectedQueueVersion: second.QueueVersion,
		ActorUserID:          "member_1",
	})
	if !IsPermissionDenied(err) {
		t.Fatalf("expected permission denied before enabling reorder, got %v", err)
	}

	_, err = svc.UpdatePermissions(context.Background(), session.JamID, sharedauth.Claims{
		UserID:       "host_1",
		Plan:         "premium",
		SessionState: sharedauth.SessionStateValid,
	}, model.PermissionUpdateRequest{CanReorderQueue: boolPtr(true)})
	if err != nil {
		t.Fatalf("enable reorder permission: %v", err)
	}

	reordered, err := svc.Reorder(session.JamID, model.ReorderQueueRequest{
		ItemIDs:              []string{second.Items[1].ItemID, second.Items[0].ItemID},
		ExpectedQueueVersion: second.QueueVersion,
		ActorUserID:          "member_1",
	})
	if err != nil {
		t.Fatalf("expected reorder success after permission enabled, got %v", err)
	}
	if reordered.QueueVersion <= second.QueueVersion {
		t.Fatalf("expected queue version increment after reorder, got %d", reordered.QueueVersion)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
