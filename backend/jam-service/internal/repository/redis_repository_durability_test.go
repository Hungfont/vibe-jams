package repository

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"video-streaming/backend/jams/internal/model"
)

func TestDurableQueueRepository_RestartRecoversSessionQueueAndIdempotency(t *testing.T) {
	t.Parallel()

	statePath := filepath.Join(t.TempDir(), "jam-state.json")
	repo, err := NewDurableQueueRepository(statePath)
	if err != nil {
		t.Fatalf("NewDurableQueueRepository() error = %v", err)
	}

	created, err := repo.CreateSession("host_1")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if _, err := repo.JoinSession(created.JamID, "member_1"); err != nil {
		t.Fatalf("JoinSession() error = %v", err)
	}

	addReq := model.AddQueueItemRequest{
		TrackID:        "trk_1",
		AddedBy:        "host_1",
		IdempotencyKey: "idem_1",
	}
	added, fromCache, err := repo.Add(created.JamID, addReq)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if fromCache {
		t.Fatal("expected first add to mutate queue")
	}

	restarted, err := NewDurableQueueRepository(statePath)
	if err != nil {
		t.Fatalf("restart NewDurableQueueRepository() error = %v", err)
	}

	session, err := restarted.SessionSnapshot(created.JamID)
	if err != nil {
		t.Fatalf("SessionSnapshot() error = %v", err)
	}
	if session.Status != model.SessionStatusActive {
		t.Fatalf("session status mismatch: got %q want %q", session.Status, model.SessionStatusActive)
	}
	if len(session.Participants) != 2 {
		t.Fatalf("participants mismatch: got %d want 2", len(session.Participants))
	}

	snapshot, err := restarted.Snapshot(created.JamID)
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}
	if snapshot.QueueVersion != added.QueueVersion {
		t.Fatalf("queue version mismatch: got %d want %d", snapshot.QueueVersion, added.QueueVersion)
	}
	if len(snapshot.Items) != 1 {
		t.Fatalf("queue items mismatch: got %d want 1", len(snapshot.Items))
	}

	retry, fromCache, err := restarted.Add(created.JamID, addReq)
	if err != nil {
		t.Fatalf("Add() retry error = %v", err)
	}
	if !fromCache {
		t.Fatal("expected idempotent retry after restart to use cached result")
	}
	if len(retry.Items) != 1 {
		t.Fatalf("idempotent retry item count mismatch: got %d want 1", len(retry.Items))
	}
}

func TestDurableQueueRepository_ConcurrentAddsRemainConsistent(t *testing.T) {
	t.Parallel()

	statePath := filepath.Join(t.TempDir(), "jam-state.json")
	repo, err := NewDurableQueueRepository(statePath)
	if err != nil {
		t.Fatalf("NewDurableQueueRepository() error = %v", err)
	}

	created, err := repo.CreateSession("host_1")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	const writers = 24
	var wg sync.WaitGroup
	errCh := make(chan error, writers)

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _, addErr := repo.Add(created.JamID, model.AddQueueItemRequest{
				TrackID:        fmt.Sprintf("trk_%d", i),
				AddedBy:        "host_1",
				IdempotencyKey: fmt.Sprintf("idem_%d", i),
			})
			if addErr != nil {
				errCh <- addErr
			}
		}(i)
	}

	wg.Wait()
	close(errCh)
	for addErr := range errCh {
		t.Fatalf("concurrent add failed: %v", addErr)
	}

	snapshot, err := repo.Snapshot(created.JamID)
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}
	if len(snapshot.Items) != writers {
		t.Fatalf("queue item count mismatch: got %d want %d", len(snapshot.Items), writers)
	}
	if snapshot.QueueVersion != writers {
		t.Fatalf("queue version mismatch: got %d want %d", snapshot.QueueVersion, writers)
	}

	restarted, err := NewDurableQueueRepository(statePath)
	if err != nil {
		t.Fatalf("restart NewDurableQueueRepository() error = %v", err)
	}
	reloaded, err := restarted.Snapshot(created.JamID)
	if err != nil {
		t.Fatalf("restarted Snapshot() error = %v", err)
	}
	if reloaded.QueueVersion != snapshot.QueueVersion {
		t.Fatalf("reloaded queue version mismatch: got %d want %d", reloaded.QueueVersion, snapshot.QueueVersion)
	}
}
