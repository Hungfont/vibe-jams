package bff

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"video-streaming/backend/api-service/internal/config"
)

func TestProxyHandler_JamRouteDelegatesAndForwardsIdentity(t *testing.T) {
	t.Parallel()

	var gotAuth string
	var gotXAuthUserID string
	jamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotXAuthUserID = r.Header.Get("X-Auth-UserId")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jamId":"jam_1"}`))
	}))
	defer jamServer.Close()

	playbackServer := httptest.NewServer(http.NotFoundHandler())
	defer playbackServer.Close()
	catalogServer := httptest.NewServer(http.NotFoundHandler())
	defer catalogServer.Close()

	h, err := NewProxyHandler(config.Config{
		JamServiceURL:      jamServer.URL,
		PlaybackServiceURL: playbackServer.URL,
		CatalogServiceURL:  catalogServer.URL,
		RTGatewayURL:       "http://localhost:8086",
		GatewayPublicURL:   "http://localhost:8085",
	})
	if err != nil {
		t.Fatalf("NewProxyHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jams/jam_1/state", nil)
	req.Header.Set("Authorization", "Bearer token-1")
	req.Header.Set("X-Auth-UserId", "user-1")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}
	if gotAuth != "Bearer token-1" {
		t.Fatalf("Authorization header mismatch: got %q", gotAuth)
	}
	if gotXAuthUserID != "user-1" {
		t.Fatalf("X-Auth-UserId mismatch: got %q", gotXAuthUserID)
	}
}

func TestProxyHandler_PlaybackAndCatalogDelegation(t *testing.T) {
	t.Parallel()

	playbackCalled := false
	catalogCalled := false

	jamServer := httptest.NewServer(http.NotFoundHandler())
	defer jamServer.Close()
	playbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		playbackCalled = true
		_, _ = w.Write([]byte(`{"accepted":true}`))
	}))
	defer playbackServer.Close()
	catalogServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		catalogCalled = true
		_, _ = w.Write([]byte(`{"trackId":"trk_1"}`))
	}))
	defer catalogServer.Close()

	h, err := NewProxyHandler(config.Config{
		JamServiceURL:      jamServer.URL,
		PlaybackServiceURL: playbackServer.URL,
		CatalogServiceURL:  catalogServer.URL,
		RTGatewayURL:       "http://localhost:8086",
		GatewayPublicURL:   "http://localhost:8085",
	})
	if err != nil {
		t.Fatalf("NewProxyHandler() error = %v", err)
	}

	recPlayback := httptest.NewRecorder()
	h.ServeHTTP(recPlayback, httptest.NewRequest(http.MethodPost, "/v1/jam/sessions/jam_1/playback/commands", nil))
	if recPlayback.Code != http.StatusOK {
		t.Fatalf("playback status mismatch: got %d want %d", recPlayback.Code, http.StatusOK)
	}

	recCatalog := httptest.NewRecorder()
	h.ServeHTTP(recCatalog, httptest.NewRequest(http.MethodGet, "/internal/v1/catalog/tracks/trk_1", nil))
	if recCatalog.Code != http.StatusOK {
		t.Fatalf("catalog status mismatch: got %d want %d", recCatalog.Code, http.StatusOK)
	}

	if !playbackCalled {
		t.Fatal("expected playback downstream to be called")
	}
	if !catalogCalled {
		t.Fatal("expected catalog downstream to be called")
	}
}

func TestProxyHandler_RealtimeWSConfig(t *testing.T) {
	t.Parallel()

	h, err := NewProxyHandler(config.Config{
		JamServiceURL:      "http://localhost:8080",
		PlaybackServiceURL: "http://localhost:8082",
		CatalogServiceURL:  "http://localhost:8083",
		RTGatewayURL:       "http://localhost:8086",
		GatewayPublicURL:   "http://localhost:8085",
	})
	if err != nil {
		t.Fatalf("NewProxyHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/bff/mvp/realtime/ws-config?sessionId=jam_1&lastSeenVersion=3", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}
	var payload struct {
		WSURL           string `json:"wsUrl"`
		SessionID       string `json:"sessionId"`
		LastSeenVersion string `json:"lastSeenVersion"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.WSURL != "ws://localhost:8085/v1/bff/mvp/realtime/ws" {
		t.Fatalf("wsUrl mismatch: got %q", payload.WSURL)
	}
	if payload.SessionID != "jam_1" {
		t.Fatalf("sessionId mismatch: got %q", payload.SessionID)
	}
	if payload.LastSeenVersion != "3" {
		t.Fatalf("lastSeenVersion mismatch: got %q", payload.LastSeenVersion)
	}
}

func TestProxyHandler_RealtimeWSProxy_RewritesToRTGatewayWSPath(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotQuery string
	rtGatewayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer rtGatewayServer.Close()

	h, err := NewProxyHandler(config.Config{
		JamServiceURL:      "http://localhost:8080",
		PlaybackServiceURL: "http://localhost:8082",
		CatalogServiceURL:  "http://localhost:8083",
		RTGatewayURL:       rtGatewayServer.URL,
		GatewayPublicURL:   "http://localhost:8085",
	})
	if err != nil {
		t.Fatalf("NewProxyHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/bff/mvp/realtime/ws?sessionId=jam_1&lastSeenVersion=7", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}
	if gotPath != "/ws" {
		t.Fatalf("expected rewritten rt-gateway path /ws, got %q", gotPath)
	}
	if gotQuery != "sessionId=jam_1&lastSeenVersion=7" {
		t.Fatalf("query mismatch: got %q", gotQuery)
	}
}
