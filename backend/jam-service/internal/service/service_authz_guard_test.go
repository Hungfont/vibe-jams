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

func TestModerationNonHostDeniedFastWithHostOnlyAndDeniedAudit(t *testing.T) {
	t.Parallel()

	repo := repository.NewRedisQueueRepository()
	pub := &kafka.InMemoryPublisher{}
	svc := New(repo, kafka.NewProducer(pub))

	session, err := svc.CreateSession("host_1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := svc.JoinSession(session.JamID, "member_1"); err != nil {
		t.Fatalf("join session: %v", err)
	}

	_, err = svc.MuteParticipant(context.Background(), session.JamID, sharedauth.Claims{
		UserID:       "member_1",
		Plan:         "free",
		SessionState: sharedauth.SessionStateValid,
	}, model.ModerationCommandRequest{TargetUserID: "host_1", Reason: "not-allowed"})
	if !IsHostOnly(err) {
		t.Fatalf("expected host-only denial, got %v", err)
	}

	snapshot, err := svc.SessionSnapshot(session.JamID)
	if err != nil {
		t.Fatalf("session snapshot: %v", err)
	}
	for _, participant := range snapshot.Participants {
		if participant.UserID == "host_1" && participant.Muted {
			t.Fatal("moderation side effect should not run on denied authorization")
		}
	}

	foundDeniedAudit := false
	for _, record := range pub.Records {
		if record.Topic != sharedkafka.TopicJamSession {
			continue
		}
		env, err := sharedevent.UnmarshalEnvelope(record.Value)
		if err != nil {
			t.Fatalf("unmarshal envelope: %v", err)
		}
		if env.EventType != "jam.policy.authorization.decided" {
			continue
		}

		var payload struct {
			Outcome     string `json:"outcome"`
			Reason      string `json:"reason"`
			ActorUserID string `json:"actorUserId"`
			ActorPlan   string `json:"actorPlan"`
		}
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}
		if payload.Outcome == "denied" {
			if payload.Reason != "host_only" {
				t.Fatalf("denied reason mismatch: got %q want host_only", payload.Reason)
			}
			if payload.ActorUserID != "member_1" || payload.ActorPlan != "free" {
				t.Fatalf("actor metadata mismatch: %+v", payload)
			}
			foundDeniedAudit = true
		}
	}

	if !foundDeniedAudit {
		t.Fatal("expected denied authorization audit event")
	}
}

func TestPolicyAuthorizationConsistencyAcrossModerationAndPermissionEntrypoints(t *testing.T) {
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

	memberClaims := sharedauth.Claims{
		UserID:       "member_1",
		Plan:         "free",
		SessionState: sharedauth.SessionStateValid,
	}

	_, moderationErr := svc.KickParticipant(context.Background(), session.JamID, memberClaims, model.ModerationCommandRequest{
		TargetUserID: "host_1",
		Reason:       "not-allowed",
	})
	if !IsHostOnly(moderationErr) {
		t.Fatalf("expected host-only moderation denial, got %v", moderationErr)
	}

	permissionErr := svc.AuthorizePermissionCommand(context.Background(), session.JamID, memberClaims, "canControlPlayback")
	if !IsHostOnly(permissionErr) {
		t.Fatalf("expected host-only permission denial, got %v", permissionErr)
	}

	hostClaims := sharedauth.Claims{
		UserID:       "host_1",
		Plan:         "premium",
		SessionState: sharedauth.SessionStateValid,
	}
	if err := svc.AuthorizePermissionCommand(context.Background(), session.JamID, hostClaims, "canControlPlayback"); err != nil {
		t.Fatalf("expected host permission authorization success, got %v", err)
	}
}
