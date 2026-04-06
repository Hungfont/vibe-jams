package repository

import (
	"errors"
	"testing"

	"video-streaming/backend/jams/internal/model"
)

func TestModerationMuteAndKick_BlockQueueActions(t *testing.T) {
	t.Parallel()

	repo := NewRedisQueueRepository()
	session, err := repo.CreateSession("host_1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := repo.JoinSession(session.JamID, "member_1"); err != nil {
		t.Fatalf("join session: %v", err)
	}

	if _, err := repo.MuteParticipant(session.JamID, "host_1", model.ModerationCommandRequest{TargetUserID: "member_1", Reason: "spam"}); err != nil {
		t.Fatalf("mute participant: %v", err)
	}

	_, _, err = repo.Add(session.JamID, model.AddQueueItemRequest{TrackID: "trk_1", AddedBy: "member_1", IdempotencyKey: "k1"})
	if !errors.Is(err, ErrModerationBlocked) {
		t.Fatalf("expected ErrModerationBlocked for muted participant, got %v", err)
	}

	if _, err := repo.KickParticipant(session.JamID, "host_1", model.ModerationCommandRequest{TargetUserID: "member_1", Reason: "abuse"}); err != nil {
		t.Fatalf("kick participant: %v", err)
	}

	_, _, err = repo.Add(session.JamID, model.AddQueueItemRequest{TrackID: "trk_2", AddedBy: "member_1", IdempotencyKey: "k2"})
	if !errors.Is(err, ErrModerationBlocked) {
		t.Fatalf("expected ErrModerationBlocked for kicked participant, got %v", err)
	}
}

func TestModerationNonHostRejected(t *testing.T) {
	t.Parallel()

	repo := NewRedisQueueRepository()
	session, err := repo.CreateSession("host_1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := repo.JoinSession(session.JamID, "member_1"); err != nil {
		t.Fatalf("join session: %v", err)
	}

	_, err = repo.MuteParticipant(session.JamID, "member_1", model.ModerationCommandRequest{TargetUserID: "host_1", Reason: "nope"})
	if !errors.Is(err, ErrHostOwnershipRequired) {
		t.Fatalf("expected ErrHostOwnershipRequired, got %v", err)
	}
}
