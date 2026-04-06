package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"video-streaming/backend/rt-gateway/internal/fanout"
	gatewaykafka "video-streaming/backend/rt-gateway/internal/kafka"
	"video-streaming/backend/rt-gateway/internal/metrics"
	"video-streaming/backend/rt-gateway/internal/snapshot"
	sharedevent "video-streaming/backend/shared/event"
)

var sessionIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

// App wires HTTP endpoints, Kafka consumer loop, and fanout processor.
type App struct {
	cfg       Config
	consumer  gatewaykafka.Consumer
	metrics   *metrics.Registry
	hub       *fanout.Hub
	processor *fanout.Processor
	hook      ModerationEventHook
}

// NewApp builds the gateway runtime graph.
func NewApp(cfg Config, consumer gatewaykafka.Consumer) *App {
	return NewAppWithModerationHook(cfg, consumer, NoopModerationEventHook{})
}

// NewAppWithModerationHook builds the gateway runtime graph with a moderation hook.
func NewAppWithModerationHook(cfg Config, consumer gatewaykafka.Consumer, hook ModerationEventHook) *App {
	registry := metrics.NewRegistry()
	hub := fanout.NewHub(cfg.FanoutBufferSize, registry)
	fetcher := snapshot.NewClient(cfg.JamServiceURL, cfg.SnapshotTimeout)
	processor := fanout.NewProcessor(hub, registry, fetcher, cfg.RecoveryMaxRetries, cfg.RecoveryBackoff)
	if hook == nil {
		hook = NoopModerationEventHook{}
	}
	return &App{
		cfg:       cfg,
		consumer:  consumer,
		metrics:   registry,
		hub:       hub,
		processor: processor,
		hook:      hook,
	}
}

// Handler returns an HTTP multiplexer for health, metrics, and websocket APIs.
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", a.handleHealth)
	mux.HandleFunc("/metrics/fanout", a.handleMetrics)
	mux.HandleFunc("/ws", a.handleWebsocket)
	return mux
}

// StartConsumer runs the fanout consume loop.
func (a *App) StartConsumer(ctx context.Context) error {
	if !a.cfg.FeatureFanoutEnabled {
		slog.Info("fanout consumer disabled by feature flag")
		<-ctx.Done()
		return ctx.Err()
	}

	return a.consumer.Start(ctx, func(consumeCtx context.Context, record gatewaykafka.Record) error {
		isModerationEvent := record.Topic == a.cfg.ModerationTopic
		if record.Topic != a.cfg.QueueTopic && record.Topic != a.cfg.PlaybackTopic && !isModerationEvent {
			return nil
		}

		envelope, err := sharedevent.UnmarshalEnvelope(record.Value)
		if err != nil {
			return fmt.Errorf("decode envelope: %w", err)
		}
		if isModerationEvent {
			if err := a.hook.HandleModerationEvent(consumeCtx, envelope); err != nil {
				return fmt.Errorf("handle moderation hook: %w", err)
			}
		}
		return a.processor.HandleEnvelope(consumeCtx, envelope)
	})
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "rt-gateway",
	})
}

func (a *App) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.metrics.Snapshot())
}

func (a *App) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	if !a.cfg.FeatureFanoutEnabled {
		http.Error(w, "realtime fanout is disabled", http.StatusServiceUnavailable)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if !isOriginAllowed(origin, a.cfg.AllowedOrigins) {
		slog.Warn("websocket origin rejected", "origin", origin)
		writeJSON(w, http.StatusForbidden, map[string]any{
			"error": map[string]string{
				"code":    "forbidden_origin",
				"message": "origin is not allowed",
			},
		})
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if !sessionIDPattern.MatchString(sessionID) {
		http.Error(w, "invalid sessionId", http.StatusBadRequest)
		return
	}

	lastSeenVersion, hasCursor, err := parseLastSeenVersion(r.URL.Query().Get("lastSeenVersion"))
	if err != nil {
		http.Error(w, "invalid lastSeenVersion", http.StatusBadRequest)
		return
	}

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	if hasCursor {
		ctx, cancel := context.WithTimeout(r.Context(), a.cfg.SnapshotTimeout)
		snapshotPayload, shouldSend, reconnectErr := a.processor.HandleReconnect(ctx, sessionID, lastSeenVersion)
		cancel()
		if reconnectErr != nil {
			slog.Warn("reconnect snapshot failed", "sessionId", sessionID, "error", reconnectErr)
		} else if shouldSend {
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if writeErr := conn.WriteMessage(websocket.TextMessage, snapshotPayload); writeErr != nil {
				return
			}
		}
	}

	subscriber := a.hub.AddSubscriber(sessionID)
	defer a.hub.RemoveSubscriber(sessionID, subscriber)

	go func() {
		for payload := range subscriber.Send {
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		}
	}()

	for {
		_, message, readErr := conn.ReadMessage()
		if readErr != nil {
			return
		}

		var command struct {
			Action string `json:"action"`
		}
		if err := json.Unmarshal(message, &command); err == nil && command.Action == "unsubscribe" {
			return
		}
	}
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return false
	}
	for _, allowed := range allowedOrigins {
		if strings.EqualFold(strings.TrimSpace(allowed), origin) {
			return true
		}
	}
	return false
}

func parseLastSeenVersion(raw string) (int64, bool, error) {
	if raw == "" {
		return 0, false, nil
	}
	version, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, err
	}
	if version < 0 {
		return 0, false, fmt.Errorf("must be >= 0")
	}
	return version, true, nil
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}
