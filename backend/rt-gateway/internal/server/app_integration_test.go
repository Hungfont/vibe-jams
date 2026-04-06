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
		ModerationTopic:      "jam.moderation.events",
		KafkaBootstrap:       "localhost:9092",
		ConsumerBackend:      "noop",
		AllowedOrigins:       []string{"http://localhost:3000"},
		RecoveryMaxRetries:   1,
		RecoveryBackoff:      5 * time.Millisecond,
		FeatureFanoutEnabled: true,
	}

	app := NewApp(cfg, kafka.NewNoopConsumer())
	testServer := httptest.NewServer(app.Handler())
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws?sessionId=jam_1"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": []string{"http://localhost:3000"}})
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

func TestWebsocketFanout_EndToEndViaConsumerLoop(t *testing.T) {
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
		ModerationTopic:      "jam.moderation.events",
		KafkaBootstrap:       "localhost:9092",
		ConsumerBackend:      "kafka",
		AllowedOrigins:       []string{"http://localhost:3000"},
		RecoveryMaxRetries:   1,
		RecoveryBackoff:      5 * time.Millisecond,
		FeatureFanoutEnabled: true,
	}

	consumer := kafka.NewInMemoryConsumer(8)
	app := NewApp(cfg, consumer)
	testServer := httptest.NewServer(app.Handler())
	defer testServer.Close()

	consumeCtx, cancelConsume := context.WithCancel(context.Background())
	defer cancelConsume()
	consumerErrCh := make(chan error, 1)
	go func() {
		consumerErrCh <- app.StartConsumer(consumeCtx)
	}()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws?sessionId=jam_1"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": []string{"http://localhost:3000"}})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	envelope, err := sharedevent.MarshalEnvelope(sharedevent.Envelope{
		EventID:          "evt-end-to-end-1",
		EventType:        "jam.queue.item.added",
		SessionID:        "jam_1",
		AggregateVersion: 1,
		OccurredAt:       time.Now().UTC(),
		Payload:          sharedevent.MustPayload(map[string]any{"trackId": "t1"}),
	}, true)
	if err != nil {
		t.Fatalf("MarshalEnvelope() error = %v", err)
	}

	if ok := consumer.Publish(kafka.Record{Topic: cfg.QueueTopic, Value: envelope}); !ok {
		t.Fatal("failed to publish test record to in-memory consumer")
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

	cancelConsume()
	if consumeErr := <-consumerErrCh; consumeErr != nil && consumeErr != context.Canceled {
		t.Fatalf("StartConsumer() error = %v", consumeErr)
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
		ModerationTopic:      "jam.moderation.events",
		KafkaBootstrap:       "localhost:9092",
		ConsumerBackend:      "noop",
		AllowedOrigins:       []string{"http://localhost:3000"},
		RecoveryMaxRetries:   1,
		RecoveryBackoff:      5 * time.Millisecond,
		FeatureFanoutEnabled: true,
	}

	app := NewApp(cfg, kafka.NewNoopConsumer())
	app.hub.SetLastVersion("jam_1", 5)

	testServer := httptest.NewServer(app.Handler())
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws?sessionId=jam_1&lastSeenVersion=2"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": []string{"http://localhost:3000"}})
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

func TestWebsocketRejectsUnknownOrigin(t *testing.T) {
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
		ModerationTopic:      "jam.moderation.events",
		KafkaBootstrap:       "localhost:9092",
		ConsumerBackend:      "noop",
		AllowedOrigins:       []string{"http://localhost:3000"},
		RecoveryMaxRetries:   1,
		RecoveryBackoff:      5 * time.Millisecond,
		FeatureFanoutEnabled: true,
	}

	app := NewApp(cfg, kafka.NewNoopConsumer())
	testServer := httptest.NewServer(app.Handler())
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws?sessionId=jam_1"
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": []string{"https://evil.example"}})
	if err == nil {
		t.Fatal("expected websocket dial failure for forbidden origin")
	}
	if resp == nil {
		t.Fatal("expected HTTP response for forbidden origin")
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status mismatch: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

type spyModerationHook struct {
	called chan sharedevent.Envelope
}

func (s *spyModerationHook) HandleModerationEvent(_ context.Context, envelope sharedevent.Envelope) error {
	s.called <- envelope
	return nil
}

func TestModerationTopicInvokesHookAndFansOut(t *testing.T) {
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
		ModerationTopic:      "jam.moderation.events",
		KafkaBootstrap:       "localhost:9092",
		ConsumerBackend:      "kafka",
		AllowedOrigins:       []string{"http://localhost:3000"},
		RecoveryMaxRetries:   1,
		RecoveryBackoff:      5 * time.Millisecond,
		FeatureFanoutEnabled: true,
	}

	consumer := kafka.NewInMemoryConsumer(8)
	hook := &spyModerationHook{called: make(chan sharedevent.Envelope, 1)}
	app := NewAppWithModerationHook(cfg, consumer, hook)
	testServer := httptest.NewServer(app.Handler())
	defer testServer.Close()

	consumeCtx, cancelConsume := context.WithCancel(context.Background())
	defer cancelConsume()
	consumerErrCh := make(chan error, 1)
	go func() {
		consumerErrCh <- app.StartConsumer(consumeCtx)
	}()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws?sessionId=jam_1"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": []string{"http://localhost:3000"}})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	envelope, err := sharedevent.MarshalEnvelope(sharedevent.Envelope{
		EventID:          "evt-mod-1",
		EventType:        "jam.moderation.muted",
		SessionID:        "jam_1",
		AggregateVersion: 1,
		OccurredAt:       time.Now().UTC(),
		Payload:          sharedevent.MustPayload(map[string]any{"action": "mute", "targetUserId": "member_1"}),
	}, true)
	if err != nil {
		t.Fatalf("MarshalEnvelope() error = %v", err)
	}

	if ok := consumer.Publish(kafka.Record{Topic: cfg.ModerationTopic, Value: envelope}); !ok {
		t.Fatal("failed to publish moderation test record")
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
	if outbound.EventType != "jam.moderation.muted" {
		t.Fatalf("event type mismatch: got %q want jam.moderation.muted", outbound.EventType)
	}

	select {
	case received := <-hook.called:
		if received.EventType != "jam.moderation.muted" {
			t.Fatalf("hook event type mismatch: got %q", received.EventType)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected moderation hook invocation")
	}

	cancelConsume()
	if consumeErr := <-consumerErrCh; consumeErr != nil && consumeErr != context.Canceled {
		t.Fatalf("StartConsumer() error = %v", consumeErr)
	}
}
