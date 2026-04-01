package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"video-streaming/backend/rt-gateway/internal/kafka"
	"video-streaming/backend/rt-gateway/internal/model"
	sharedevent "video-streaming/backend/shared/event"
)

func TestWebsocketSubscriptionAndFanout(t *testing.T) {
	t.Parallel()

	cfg := Config{
		ServerHost:           "127.0.0.1",
		ServerPort:           0,
		ReadHeaderTimeout:    5 * time.Second,
		IdleTimeout:          30 * time.Second,
		ShutdownTimeout:      5 * time.Second,
		JamServiceURL:        "http://localhost:8080",
		SnapshotTimeout:      500 * time.Millisecond,
		FanoutBufferSize:     8,
		ConsumerGroupID:      "rt-gateway-fanout",
		QueueTopic:           "jam.queue.events",
		PlaybackTopic:        "jam.playback.events",
		RecoveryMaxRetries:   1,
		RecoveryBackoff:      5 * time.Millisecond,
		FeatureFanoutEnabled: true,
	}

	app := NewApp(cfg, kafka.NewNoopConsumer())
	testServer := httptest.NewServer(app.Handler())
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws?sessionId=jam_1"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	err = app.processor.HandleEnvelope(context.Background(), sharedevent.Envelope{
		EventID:          "evt-1",
		EventType:        "jam.queue.item.added",
		SessionID:        "jam_1",
		AggregateVersion: 1,
		OccurredAt:       time.Now().UTC(),
		Payload:          sharedevent.MustPayload(map[string]any{"trackId": "t1"}),
	})
	if err != nil {
		t.Fatalf("HandleEnvelope() error = %v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read websocket message: %v", err)
	}

	var outbound model.OutboundEvent
	if err := json.Unmarshal(message, &outbound); err != nil {
		t.Fatalf("decode outbound message: %v", err)
	}
	if outbound.SessionID != "jam_1" {
		t.Fatalf("session mismatch: got %q want jam_1", outbound.SessionID)
	}
	if outbound.AggregateVersion != 1 {
		t.Fatalf("aggregateVersion mismatch: got %d want 1", outbound.AggregateVersion)
	}
}

func TestReconnectStaleCursorReceivesSnapshotFallback(t *testing.T) {
	t.Parallel()

	snapshotServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"session":{"jamId":"jam_1","status":"active","hostUserId":"host_1","sessionVersion":5,"participants":[{"userId":"host_1","role":"host"}]},"queue":{"jamId":"jam_1","queueVersion":5,"items":[]},"aggregateVersion":5}`))
	}))
	defer snapshotServer.Close()

	cfg := Config{
		ServerHost:           "127.0.0.1",
		ServerPort:           0,
		ReadHeaderTimeout:    5 * time.Second,
		IdleTimeout:          30 * time.Second,
		ShutdownTimeout:      5 * time.Second,
		JamServiceURL:        snapshotServer.URL,
		SnapshotTimeout:      1 * time.Second,
		FanoutBufferSize:     8,
		ConsumerGroupID:      "rt-gateway-fanout",
		QueueTopic:           "jam.queue.events",
		PlaybackTopic:        "jam.playback.events",
		RecoveryMaxRetries:   1,
		RecoveryBackoff:      5 * time.Millisecond,
		FeatureFanoutEnabled: true,
	}

	app := NewApp(cfg, kafka.NewNoopConsumer())
	app.hub.SetLastVersion("jam_1", 5)

	testServer := httptest.NewServer(app.Handler())
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws?sessionId=jam_1&lastSeenVersion=2"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read snapshot message: %v", err)
	}

	var outbound model.OutboundEvent
	if err := json.Unmarshal(message, &outbound); err != nil {
		t.Fatalf("decode outbound snapshot: %v", err)
	}
	if outbound.EventType != "jam.session.snapshot" {
		t.Fatalf("eventType mismatch: got %q want jam.session.snapshot", outbound.EventType)
	}
	if !outbound.Recovery {
		t.Fatal("expected recovery flag for snapshot fallback")
	}
}
