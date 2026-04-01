package repository

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"video-streaming/backend/playback-service/internal/model"
)

func TestDurablePlaybackRepository_RestartRecoversTransitions(t *testing.T) {
	t.Parallel()

	statePath := filepath.Join(t.TempDir(), "playback-state.json")
	repo, err := NewDurablePlaybackRepository(statePath)
	if err != nil {
		t.Fatalf("NewDurablePlaybackRepository() error = %v", err)
	}
	if err := repo.SeedSession("jam_1", "host_1", 7); err != nil {
		t.Fatalf("SeedSession() error = %v", err)
	}

	first, err := repo.ApplyCommand("jam_1", model.CommandPlay, 0, "host_1", "evt_1")
	if err != nil {
		t.Fatalf("ApplyCommand(play) error = %v", err)
	}
	if first.AggregateVersion != 1 {
		t.Fatalf("aggregateVersion mismatch: got %d want 1", first.AggregateVersion)
	}

	restarted, err := NewDurablePlaybackRepository(statePath)
	if err != nil {
		t.Fatalf("restart NewDurablePlaybackRepository() error = %v", err)
	}

	queueVersion, err := restarted.QueueVersion("jam_1")
	if err != nil {
		t.Fatalf("QueueVersion() error = %v", err)
	}
	if queueVersion != 7 {
		t.Fatalf("queue version mismatch: got %d want 7", queueVersion)
	}

	second, err := restarted.ApplyCommand("jam_1", model.CommandPause, 0, "host_1", "evt_2")
	if err != nil {
		t.Fatalf("ApplyCommand(pause) error = %v", err)
	}
	if second.AggregateVersion != 2 {
		t.Fatalf("aggregateVersion mismatch after restart: got %d want 2", second.AggregateVersion)
	}
	if second.State != "paused" {
		t.Fatalf("state mismatch after restart: got %q want paused", second.State)
	}
}

func TestDurablePlaybackRepository_ConcurrentCommandsRemainConsistent(t *testing.T) {
	t.Parallel()

	statePath := filepath.Join(t.TempDir(), "playback-state.json")
	repo, err := NewDurablePlaybackRepository(statePath)
	if err != nil {
		t.Fatalf("NewDurablePlaybackRepository() error = %v", err)
	}
	if err := repo.SeedSession("jam_1", "host_1", 9); err != nil {
		t.Fatalf("SeedSession() error = %v", err)
	}

	const writers = 20
	var wg sync.WaitGroup
	errCh := make(chan error, writers)

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, applyErr := repo.ApplyCommand("jam_1", model.CommandSeek, int64(i*100), "host_1", fmt.Sprintf("evt_%d", i))
			if applyErr != nil {
				errCh <- applyErr
			}
		}(i)
	}

	wg.Wait()
	close(errCh)
	for applyErr := range errCh {
		t.Fatalf("concurrent command failed: %v", applyErr)
	}

	restarted, err := NewDurablePlaybackRepository(statePath)
	if err != nil {
		t.Fatalf("restart NewDurablePlaybackRepository() error = %v", err)
	}

	finalTransition, err := restarted.ApplyCommand("jam_1", model.CommandPause, 0, "host_1", "evt_final")
	if err != nil {
		t.Fatalf("ApplyCommand(final) error = %v", err)
	}
	if finalTransition.AggregateVersion != writers+1 {
		t.Fatalf("aggregateVersion mismatch: got %d want %d", finalTransition.AggregateVersion, writers+1)
	}
}
