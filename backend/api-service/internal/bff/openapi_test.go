package bff

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"video-streaming/backend/api-service/internal/config"
)

func TestOpenAPISpec_IncludesDelegatedBFFRouteFamilies(t *testing.T) {
	t.Parallel()

	spec := openAPISpec()
	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatal("openapi paths map missing")
	}

	requiredPaths := []string{
		"/v1/bff/mvp/sessions/{sessionId}/orchestration",
		"/api/v1/jams/{jamId}/state",
		"/v1/jam/sessions/{jamId}/playback/commands",
		"/internal/v1/catalog/tracks/{trackId}",
		"/v1/bff/mvp/realtime/ws-config",
		"/v1/bff/mvp/realtime/ws",
	}
	for _, route := range requiredPaths {
		if _, exists := paths[route]; !exists {
			t.Fatalf("expected openapi to include delegated path %q", route)
		}
	}
}

func TestRouter_SwaggerAndOpenAPIRoutes_AreServed(t *testing.T) {
	t.Parallel()

	router, err := NewRouter(config.Config{
		ServerHost:         "0.0.0.0",
		ServerPort:         8084,
		ReadHeaderTimeout:  5 * time.Second,
		IdleTimeout:        30 * time.Second,
		ShutdownTimeout:    10 * time.Second,
		FeatureBFFEnabled:  true,
		JamServiceURL:      "http://localhost:8080",
		PlaybackServiceURL: "http://localhost:8082",
		CatalogServiceURL:  "http://localhost:8083",
		RTGatewayURL:       "http://localhost:8086",
		GatewayPublicURL:   "http://localhost:8085",
		JamTimeout:         1200 * time.Millisecond,
		PlaybackTimeout:    1000 * time.Millisecond,
		CatalogTimeout:     800 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("build router: %v", err)
	}

	uiReq := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	uiRec := httptest.NewRecorder()
	router.ServeHTTP(uiRec, uiReq)
	if uiRec.Code != http.StatusOK {
		t.Fatalf("swagger status mismatch: got %d want %d", uiRec.Code, http.StatusOK)
	}

	openReq := httptest.NewRequest(http.MethodGet, "/swagger/openapi.json", nil)
	openRec := httptest.NewRecorder()
	router.ServeHTTP(openRec, openReq)
	if openRec.Code != http.StatusOK {
		t.Fatalf("openapi status mismatch: got %d want %d", openRec.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.NewDecoder(openRec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode openapi payload: %v", err)
	}
	if payload["openapi"] == nil {
		t.Fatal("openapi version must be present")
	}
}
