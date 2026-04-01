package bff

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOrchestrationSuccessAcrossDependencies(t *testing.T) {
	t.Parallel()

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"userId":"host_1","plan":"premium","sessionState":"valid"}`))
	}))
	defer authServer.Close()

	jamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"session":{"jamId":"jam_1","status":"active","hostUserId":"host_1","participants":[{"userId":"host_1","role":"host"}],"sessionVersion":7},"queue":{"jamId":"jam_1","queueVersion":7,"items":[]},"aggregateVersion":7}`))
	}))
	defer jamServer.Close()

	playbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"accepted":true}`))
	}))
	defer playbackServer.Close()

	catalogServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"trackId":"trk_1","isPlayable":true,"title":"Song"}`))
	}))
	defer catalogServer.Close()

	service := NewService(
		NewHTTPAuthClient(authServer.URL, time.Second),
		NewHTTPJamClient(jamServer.URL, time.Second),
		NewHTTPPlaybackClient(playbackServer.URL, time.Second),
		NewHTTPCatalogClient(catalogServer.URL, time.Second),
		true,
	)
	h := NewHandler(service)

	body := []byte(`{"trackId":"trk_1","playbackCommand":{"command":"pause","clientEventId":"evt_1","expectedQueueVersion":7}}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}

	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			Claims struct {
				UserID string `json:"userId"`
			} `json:"claims"`
			Partial  bool `json:"partial"`
			Playback struct {
				Accepted bool `json:"accepted"`
			} `json:"playback"`
			Track struct {
				TrackID string `json:"trackId"`
			} `json:"track"`
			DependencyStatuses map[string]string `json:"dependencyStatuses"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.Success {
		t.Fatal("expected success envelope")
	}
	if envelope.Data.Claims.UserID != "host_1" {
		t.Fatalf("claims propagation mismatch: got %s", envelope.Data.Claims.UserID)
	}
	if envelope.Data.Partial {
		t.Fatal("expected non-partial response")
	}
	if !envelope.Data.Playback.Accepted {
		t.Fatal("expected playback accepted response")
	}
	if envelope.Data.Track.TrackID != "trk_1" {
		t.Fatalf("track mismatch: %s", envelope.Data.Track.TrackID)
	}
}

func TestOrchestrationTimeoutNormalizedForRequiredDependency(t *testing.T) {
	t.Parallel()

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"userId":"host_1","plan":"premium","sessionState":"valid"}`))
	}))
	defer authServer.Close()

	jamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(80 * time.Millisecond)
		_, _ = w.Write([]byte(`{"session":{},"queue":{},"aggregateVersion":0}`))
	}))
	defer jamServer.Close()

	service := NewService(
		NewHTTPAuthClient(authServer.URL, time.Second),
		NewHTTPJamClient(jamServer.URL, 10*time.Millisecond),
		NewHTTPPlaybackClient("http://127.0.0.1:0", 10*time.Millisecond),
		NewHTTPCatalogClient("http://127.0.0.1:0", 10*time.Millisecond),
		true,
	)
	h := NewHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var envelope struct {
		Success bool `json:"success"`
		Error   struct {
			Code       string `json:"code"`
			Dependency string `json:"dependency"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Error.Code != "dependency_timeout" || envelope.Error.Dependency != "jam" {
		t.Fatalf("unexpected error mapping: %+v", envelope.Error)
	}
}

func TestOrchestrationOptionalFailureReturnsPartial(t *testing.T) {
	t.Parallel()

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"userId":"host_1","plan":"premium","sessionState":"valid"}`))
	}))
	defer authServer.Close()

	jamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"session":{"jamId":"jam_1","status":"active","hostUserId":"host_1","participants":[],"sessionVersion":1},"queue":{"jamId":"jam_1","queueVersion":1,"items":[]},"aggregateVersion":1}`))
	}))
	defer jamServer.Close()

	playbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":{"code":"dependency_unavailable","message":"playback unavailable"}}`))
	}))
	defer playbackServer.Close()

	catalogServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"trackId":"trk_1","isPlayable":true}`))
	}))
	defer catalogServer.Close()

	service := NewService(
		NewHTTPAuthClient(authServer.URL, time.Second),
		NewHTTPJamClient(jamServer.URL, time.Second),
		NewHTTPPlaybackClient(playbackServer.URL, time.Second),
		NewHTTPCatalogClient(catalogServer.URL, time.Second),
		true,
	)
	h := NewHandler(service)

	body := []byte(`{"trackId":"trk_1","playbackCommand":{"command":"pause","clientEventId":"evt_1","expectedQueueVersion":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}

	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			Partial bool `json:"partial"`
			Issues  []struct {
				Dependency string `json:"dependency"`
				Code       string `json:"code"`
			} `json:"issues"`
			DependencyStatuses map[string]string `json:"dependencyStatuses"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.Success || !envelope.Data.Partial {
		t.Fatalf("expected partial success, got %+v", envelope)
	}
	if len(envelope.Data.Issues) == 0 || envelope.Data.Issues[0].Dependency != "playback" {
		t.Fatalf("expected playback issue, got %+v", envelope.Data.Issues)
	}
	if envelope.Data.DependencyStatuses["playback"] != "degraded" {
		t.Fatalf("expected degraded playback status, got %+v", envelope.Data.DependencyStatuses)
	}
}
