package fanout

import (
	"context"
	"testing"
	"time"

	"video-streaming/backend/rt-gateway/internal/metrics"
	"video-streaming/backend/rt-gateway/internal/model"
	sharedevent "video-streaming/backend/shared/event"
)

func TestLoadScenario_FanoutP95LatencyUnderTarget(t *testing.T) {
	t.Parallel()

	registry := metrics.NewRegistry()
	hub := NewHub(1024, registry)
	fetcher := &stubSnapshotFetcher{
		snapshot: model.SessionStateSnapshot{
			Session:          model.SessionSnapshot{JamID: "jam_load", SessionVersion: 1, Status: "active"},
			Queue:            model.QueueSnapshot{JamID: "jam_load", QueueVersion: 1},
			AggregateVersion: 1,
		},
	}
	processor := NewProcessor(hub, registry, fetcher, 1, 5*time.Millisecond)
	subscriber := hub.AddSubscriber("jam_load")
	defer hub.RemoveSubscriber("jam_load", subscriber)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range subscriber.Send {
		}
	}()

	for version := int64(1); version <= 1000; version++ {
		err := processor.HandleEnvelope(context.Background(), sharedevent.Envelope{
			EventID:          "evt-load",
			EventType:        "jam.queue.item.added",
			SessionID:        "jam_load",
			AggregateVersion: version,
			OccurredAt:       time.Now().UTC(),
			Payload:          sharedevent.MustPayload(map[string]any{"trackId": "t"}),
		})
		if err != nil {
			t.Fatalf("HandleEnvelope() error at version %d: %v", version, err)
		}
	}

	snapshot := registry.Snapshot()
	if snapshot.P95FanoutLatencyMS > 50 {
		t.Fatalf("p95 fanout latency too high: got %.2fms want <= 50ms", snapshot.P95FanoutLatencyMS)
	}

	hub.RemoveSubscriber("jam_load", subscriber)
	<-done
}
