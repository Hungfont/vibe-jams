package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"video-streaming/backend/jams/internal/kafka"
	"video-streaming/backend/jams/internal/model"
	"video-streaming/backend/jams/internal/repository"
	sharedauth "video-streaming/backend/shared/auth"
	sharedevent "video-streaming/backend/shared/event"
	sharedkafka "video-streaming/backend/shared/kafka"
)

func TestModerationPublishesAuditEvent(t *testing.T) {
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
		UserID:       "host_1",
		Plan:         "premium",
		SessionState: sharedauth.SessionStateValid,
		Scope:        []string{"jam:moderate"},
	}, model.ModerationCommandRequest{
		TargetUserID: "member_1",
		Reason:       "spam",
	})
	if err != nil {
		t.Fatalf("mute participant: %v", err)
	}

	found := false
	for _, record := range pub.Records {
		if record.Topic != sharedkafka.TopicJamModeration {
			continue
		}
		env, err := sharedevent.UnmarshalEnvelope(record.Value)
		if err != nil {
			t.Fatalf("unmarshal envelope: %v", err)
		}
		if env.EventType != "jam.moderation.muted" {
			continue
		}
		found = true
	}

	if !found {
		t.Fatal("expected jam.moderation.muted event on moderation topic")
	}

	foundPolicyDecision := false
	for _, record := range pub.Records {
		if record.Topic != sharedkafka.TopicJamSession {
			continue
		}
		env, err := sharedevent.UnmarshalEnvelope(record.Value)
		if err != nil {
			t.Fatalf("unmarshal policy envelope: %v", err)
		}
		if env.EventType != "jam.policy.authorization.decided" {
			continue
		}

		var payload struct {
			Outcome     string   `json:"outcome"`
			ActorUserID string   `json:"actorUserId"`
			ActorPlan   string   `json:"actorPlan"`
			ActorScope  []string `json:"actorScope"`
		}
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			t.Fatalf("unmarshal policy payload: %v", err)
		}
		if payload.Outcome != "accepted" {
			t.Fatalf("policy outcome mismatch: got %q want accepted", payload.Outcome)
		}
		if payload.ActorUserID != "host_1" || payload.ActorPlan != "premium" {
			t.Fatalf("actor metadata mismatch: %+v", payload)
		}
		if len(payload.ActorScope) != 1 || payload.ActorScope[0] != "jam:moderate" {
			t.Fatalf("actor scope mismatch: %+v", payload.ActorScope)
		}
		foundPolicyDecision = true
	}
	if !foundPolicyDecision {
		t.Fatal("expected jam.policy.authorization.decided event on session topic")
	}
}

func TestMutedParticipantBlockedByService(t *testing.T) {
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
	if _, err := svc.MuteParticipant(context.Background(), session.JamID, sharedauth.Claims{UserID: "host_1", Plan: "premium", SessionState: sharedauth.SessionStateValid}, model.ModerationCommandRequest{TargetUserID: "member_1", Reason: "spam"}); err != nil {
		t.Fatalf("mute participant: %v", err)
	}

	_, _, err = svc.Add(session.JamID, model.AddQueueItemRequest{TrackID: "trk_1", AddedBy: "member_1", IdempotencyKey: "k1"})
	if !IsModerationBlocked(err) {
		t.Fatalf("expected moderation blocked error, got %v", err)
	}

	if !errors.Is(err, ErrModerationBlocked) {
		t.Fatalf("expected service ErrModerationBlocked, got %v", err)
	}
}
