package repository

import (
	"errors"
	"sync"
	"testing"

	"video-streaming/backend/jams/internal/model"
)

func TestSessionLifecycle_CreateJoinLeaveEnd(t *testing.T) {
	t.Parallel()

	repo := NewRedisQueueRepository()

	created, err := repo.CreateSession("host_1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if created.Status != model.SessionStatusActive {
		t.Fatalf("status mismatch: got %q want %q", created.Status, model.SessionStatusActive)
	}
	if created.HostUserID != "host_1" {
		t.Fatalf("host mismatch: got %q", created.HostUserID)
	}

	joined, err := repo.JoinSession(created.JamID, "member_1")
	if err != nil {
		t.Fatalf("join session: %v", err)
	}
	if len(joined.Participants) != 2 {
		t.Fatalf("participants mismatch: got %d want 2", len(joined.Participants))
	}

	left, err := repo.LeaveSession(created.JamID, "member_1")
	if err != nil {
		t.Fatalf("leave session: %v", err)
	}
	if len(left.Participants) != 1 {
		t.Fatalf("participants mismatch after leave: got %d want 1", len(left.Participants))
	}

	ended, err := repo.EndSession(created.JamID, "host_1")
	if err != nil {
		t.Fatalf("end session: %v", err)
	}
	if ended.Status != model.SessionStatusEnded {
		t.Fatalf("status mismatch: got %q want %q", ended.Status, model.SessionStatusEnded)
	}
	if ended.EndCause != "host_request" {
		t.Fatalf("end cause mismatch: got %q want host_request", ended.EndCause)
	}
}

func TestSessionLifecycle_HostLeaveEndsSession(t *testing.T) {
	t.Parallel()

	repo := NewRedisQueueRepository()
	created, err := repo.CreateSession("host_1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	ended, err := repo.LeaveSession(created.JamID, "host_1")
	if err != nil {
		t.Fatalf("host leave: %v", err)
	}
	if ended.Status != model.SessionStatusEnded {
		t.Fatalf("status mismatch: got %q want %q", ended.Status, model.SessionStatusEnded)
	}
	if ended.EndCause != "host_leave" {
		t.Fatalf("end cause mismatch: got %q want host_leave", ended.EndCause)
	}
}

func TestSessionLifecycle_NonHostCannotEnd(t *testing.T) {
	t.Parallel()

	repo := NewRedisQueueRepository()
	created, err := repo.CreateSession("host_1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := repo.JoinSession(created.JamID, "member_1"); err != nil {
		t.Fatalf("join session: %v", err)
	}

	_, err = repo.EndSession(created.JamID, "member_1")
	if !errors.Is(err, ErrHostOwnershipRequired) {
		t.Fatalf("expected ErrHostOwnershipRequired, got %v", err)
	}
}

func TestSessionLifecycle_ConcurrentJoinAndLeave(t *testing.T) {
	t.Parallel()

	repo := NewRedisQueueRepository()
	created, err := repo.CreateSession("host_1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			userID := "user_" + string(rune('a'+i))
			_, _ = repo.JoinSession(created.JamID, userID)
			_, _ = repo.LeaveSession(created.JamID, userID)
		}(i)
	}
	wg.Wait()

	session, err := repo.SessionSnapshot(created.JamID)
	if err != nil {
		t.Fatalf("session snapshot: %v", err)
	}
	if session.Status != model.SessionStatusActive {
		t.Fatalf("status mismatch: got %q want %q", session.Status, model.SessionStatusActive)
	}
	if len(session.Participants) != 1 {
		t.Fatalf("participants mismatch after concurrent ops: got %d want 1", len(session.Participants))
	}
}
