package fanout

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"video-streaming/backend/rt-gateway/internal/metrics"
	"video-streaming/backend/rt-gateway/internal/model"
	sharedevent "video-streaming/backend/shared/event"
)

type stubSnapshotFetcher struct {
	snapshot model.SessionStateSnapshot
	err      error
	calls    int
}

func (s *stubSnapshotFetcher) FetchSessionState(_ context.Context, _ string) (model.SessionStateSnapshot, error) {
	s.calls++
	if s.err != nil {
		return model.SessionStateSnapshot{}, s.err
	}
	return s.snapshot, nil
}

func TestHandleEnvelopeSequentialAndDuplicate(t *testing.T) {
	t.Parallel()

	registry := metrics.NewRegistry()
	hub := NewHub(8, registry)
	fetcher := &stubSnapshotFetcher{}
	processor := NewProcessor(hub, registry, fetcher, 1, 10*time.Millisecond)
	subscriber := hub.AddSubscriber("jam_1")
	defer hub.RemoveSubscriber("jam_1", subscriber)

	envelope := sharedevent.Envelope{
		EventID:          "evt-1",
		EventType:        "jam.queue.item.added",
		SessionID:        "jam_1",
		AggregateVersion: 1,
		OccurredAt:       time.Now().UTC(),
		Payload:          sharedevent.MustPayload(map[string]any{"trackId": "t1"}),
	}
	if err := processor.HandleEnvelope(context.Background(), envelope); err != nil {
		t.Fatalf("HandleEnvelope() error = %v", err)
	}

	select {
	case message := <-subscriber.Send:
		var outbound model.OutboundEvent
		if err := json.Unmarshal(message, &outbound); err != nil {
			t.Fatalf("decode outbound: %v", err)
		}
		if outbound.AggregateVersion != 1 {
			t.Fatalf("aggregateVersion mismatch: got %d want 1", outbound.AggregateVersion)
		}
	default:
		t.Fatal("expected broadcast message")
	}

	if err := processor.HandleEnvelope(context.Background(), envelope); err != nil {
		t.Fatalf("duplicate HandleEnvelope() error = %v", err)
	}
	metricSnapshot := registry.Snapshot()
	if metricSnapshot.DuplicateCount != 1 {
		t.Fatalf("duplicate count mismatch: got %d want 1", metricSnapshot.DuplicateCount)
	}
}

func TestHandleEnvelopeGapTriggersRecovery(t *testing.T) {
	t.Parallel()

	registry := metrics.NewRegistry()
	hub := NewHub(8, registry)
	fetcher := &stubSnapshotFetcher{
		snapshot: model.SessionStateSnapshot{
			Session:          model.SessionSnapshot{JamID: "jam_1", SessionVersion: 2, Status: "active"},
			Queue:            model.QueueSnapshot{JamID: "jam_1", QueueVersion: 2},
			AggregateVersion: 2,
		},
	}
	processor := NewProcessor(hub, registry, fetcher, 1, 10*time.Millisecond)
	subscriber := hub.AddSubscriber("jam_1")
	defer hub.RemoveSubscriber("jam_1", subscriber)

	hub.SetLastVersion("jam_1", 1)
	incoming := sharedevent.Envelope{
		EventID:          "evt-3",
		EventType:        "jam.playback.updated",
		SessionID:        "jam_1",
		AggregateVersion: 3,
		OccurredAt:       time.Now().UTC(),
		Payload:          sharedevent.MustPayload(map[string]any{"state": "play"}),
	}
	if err := processor.HandleEnvelope(context.Background(), incoming); err != nil {
		t.Fatalf("HandleEnvelope() error = %v", err)
	}
	if fetcher.calls == 0 {
		t.Fatal("expected snapshot fetch on gap detection")
	}

	first := <-subscriber.Send
	second := <-subscriber.Send

	var firstOutbound model.OutboundEvent
	if err := json.Unmarshal(first, &firstOutbound); err != nil {
		t.Fatalf("decode first outbound: %v", err)
	}
	if firstOutbound.EventType != "jam.session.snapshot" {
		t.Fatalf("expected snapshot event first, got %q", firstOutbound.EventType)
	}

	var secondOutbound model.OutboundEvent
	if err := json.Unmarshal(second, &secondOutbound); err != nil {
		t.Fatalf("decode second outbound: %v", err)
	}
	if secondOutbound.AggregateVersion != 3 {
		t.Fatalf("expected resumed event version 3, got %d", secondOutbound.AggregateVersion)
	}
}

func TestHandleReconnectSnapshotFallback(t *testing.T) {
	t.Parallel()

	registry := metrics.NewRegistry()
	hub := NewHub(8, registry)
	fetcher := &stubSnapshotFetcher{
		snapshot: model.SessionStateSnapshot{
			Session:          model.SessionSnapshot{JamID: "jam_1", SessionVersion: 5, Status: "active"},
			Queue:            model.QueueSnapshot{JamID: "jam_1", QueueVersion: 5},
			AggregateVersion: 5,
		},
	}
	processor := NewProcessor(hub, registry, fetcher, 0, 10*time.Millisecond)
	hub.SetLastVersion("jam_1", 5)

	payload, shouldSend, err := processor.HandleReconnect(context.Background(), "jam_1", 2)
	if err != nil {
		t.Fatalf("HandleReconnect() error = %v", err)
	}
	if !shouldSend {
		t.Fatal("expected snapshot fallback for stale cursor")
	}

	var outbound model.OutboundEvent
	if err := json.Unmarshal(payload, &outbound); err != nil {
		t.Fatalf("decode outbound payload: %v", err)
	}
	if outbound.EventType != "jam.session.snapshot" {
		t.Fatalf("eventType mismatch: got %q", outbound.EventType)
	}

	_, shouldSend, err = processor.HandleReconnect(context.Background(), "jam_1", 5)
	if err != nil {
		t.Fatalf("HandleReconnect() same version error = %v", err)
	}
	if shouldSend {
		t.Fatal("did not expect snapshot for current cursor")
	}
}

func TestHandleEnvelopeGapFailure(t *testing.T) {
	t.Parallel()

	registry := metrics.NewRegistry()
	hub := NewHub(8, registry)
	fetcher := &stubSnapshotFetcher{err: errors.New("downstream unavailable")}
	processor := NewProcessor(hub, registry, fetcher, 0, 5*time.Millisecond)
	hub.SetLastVersion("jam_1", 1)

	err := processor.HandleEnvelope(context.Background(), sharedevent.Envelope{
		EventID:          "evt-3",
		EventType:        "jam.queue.item.added",
		SessionID:        "jam_1",
		AggregateVersion: 3,
		OccurredAt:       time.Now().UTC(),
		Payload:          sharedevent.MustPayload(map[string]any{"trackId": "x"}),
	})
	if err == nil {
		t.Fatal("expected recovery error")
	}

	metricSnapshot := registry.Snapshot()
	if metricSnapshot.RecoveryFailureCount == 0 {
		t.Fatal("expected recovery failure metric")
	}
}
