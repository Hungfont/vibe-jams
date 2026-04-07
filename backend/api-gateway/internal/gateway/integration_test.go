package gateway_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"video-streaming/backend/api-gateway/internal/config"
	"video-streaming/backend/api-gateway/internal/gateway"
)

func buildTestRouter(t *testing.T, authHandler, apiHandler http.HandlerFunc) (http.Handler, *httptest.Server, *httptest.Server) {
	t.Helper()
	authServer := httptest.NewServer(authHandler)
	apiServer := httptest.NewServer(apiHandler)
	t.Cleanup(func() {
		authServer.Close()
		apiServer.Close()
	})
	cfg := config.Config{
		ServerPort:      8085,
		AuthServiceURL:  authServer.URL,
		APIServiceURL:   apiServer.URL,
		AuthTimeout:     500_000_000, // 500ms
		UpstreamTimeout: 2_000_000_000,
	}
	router, err := gateway.NewRouter(cfg)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return router, authServer, apiServer
}

func TestIntegration_ValidToken_ProxiesToAPIService(t *testing.T) {
	t.Parallel()

	authCalled := false
	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/v1/auth/validate" {
			authCalled = true
			_, _ = w.Write([]byte(`{"userId":"user_1","plan":"premium","sessionState":"valid","scope":["jam:read"]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	apiCalled := false
	var receivedUserID string
	var receivedAuthHeader string
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		receivedUserID = r.Header.Get("X-Auth-UserId")
		receivedAuthHeader = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"success":true,"data":{}}`))
	})

	router, _, _ := buildTestRouter(t, authHandler, apiHandler)

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !authCalled {
		t.Fatal("expected auth-service validate to be called")
	}
	if !apiCalled {
		t.Fatal("expected api-service to be called")
	}
	if receivedUserID != "user_1" {
		t.Fatalf("X-Auth-UserId not forwarded: got %q", receivedUserID)
	}
	if receivedAuthHeader != "Bearer valid-token" {
		t.Fatalf("Authorization header should be preserved for downstream compatibility: got %q", receivedAuthHeader)
	}
}

func TestIntegration_CookieFallback_ProxiesToAPIService(t *testing.T) {
	t.Parallel()

	authCalled := false
	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/v1/auth/validate" {
			authCalled = true
			if r.Header.Get("Authorization") != "Bearer cookie-token" {
				t.Fatalf("validate should use token resolved from cookie: got %q", r.Header.Get("Authorization"))
			}
			_, _ = w.Write([]byte(`{"userId":"user_cookie","plan":"premium","sessionState":"valid"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	apiCalled := false
	var receivedAuthHeader string
	router, _, _ := buildTestRouter(t, authHandler, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		receivedAuthHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/bff/mvp/realtime/ws-config?sessionId=jam_1", nil)
	req.Header.Set("Cookie", "auth_token=cookie-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}
	if !authCalled {
		t.Fatal("expected auth-service validate to be called")
	}
	if !apiCalled {
		t.Fatal("expected api-service to be called")
	}
	if receivedAuthHeader != "Bearer cookie-token" {
		t.Fatalf("api-service should receive Authorization derived from cookie fallback: got %q", receivedAuthHeader)
	}
}

func TestIntegration_InvalidToken_GatewayRejects401(t *testing.T) {
	t.Parallel()

	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/v1/auth/validate" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"invalid_token","message":"invalid"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	apiCalled := false
	router, _, _ := buildTestRouter(t, authHandler, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		apiCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if apiCalled {
		t.Fatal("api-service must not be called for invalid token")
	}
	assertErrorCode(t, rec, "invalid_token")
}

func TestIntegration_RevokedSession_GatewayRejects401(t *testing.T) {
	t.Parallel()

	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/v1/auth/validate" {
			_, _ = w.Write([]byte(`{"userId":"user_1","plan":"premium","sessionState":"invalid"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	router, _, _ := buildTestRouter(t, authHandler, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer revoked-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorCode(t, rec, "invalid_token")
}

func TestIntegration_PublicAuthRoute_BypassesValidation(t *testing.T) {
	t.Parallel()

	validateCalled := false
	loginCalled := false
	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/v1/auth/validate" {
			validateCalled = true
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/v1/auth/login" {
			loginCalled = true
			_, _ = w.Write([]byte(`{"accessToken":"tok","tokenType":"Bearer"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	router, _, _ := buildTestRouter(t, authHandler, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	// No Authorization header — public route must pass through without validation.
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if validateCalled {
		t.Fatal("validate must NOT be called for public auth routes")
	}
	if !loginCalled {
		t.Fatal("login endpoint on auth-service must be called")
	}
}

func TestIntegration_SwaggerUIRoute_IsPublicAndServed(t *testing.T) {
	t.Parallel()

	router, _, _ := buildTestRouter(
		t,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }),
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }),
	)

	req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct == "" {
		t.Fatal("swagger route must set content-type")
	}
}

func TestIntegration_OpenAPIJSONRoute_IsPublicAndServed(t *testing.T) {
	t.Parallel()

	router, _, _ := buildTestRouter(
		t,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }),
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }),
	)

	req := httptest.NewRequest(http.MethodGet, "/swagger/openapi.json", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}
	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode openapi response: %v", err)
	}
	if payload["openapi"] == nil {
		t.Fatal("openapi document must include openapi version field")
	}
	paths, ok := payload["paths"].(map[string]any)
	if !ok {
		t.Fatal("openapi paths must be present")
	}
	required := []string{
		"/healthz",
		"/v1/auth/login",
		"/v1/bff/mvp/realtime/ws-config",
		"/v1/bff/mvp/realtime/ws",
		"/v1/bff/mvp/sessions/{sessionId}/orchestration",
	}
	for _, route := range required {
		if _, exists := paths[route]; !exists {
			t.Fatalf("expected gateway openapi to include route %q", route)
		}
	}
}

func TestIntegration_SpoofedXAuthHeader_IsStripped(t *testing.T) {
	t.Parallel()

	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/v1/auth/validate" {
			_, _ = w.Write([]byte(`{"userId":"real_user","plan":"free","sessionState":"valid"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	var receivedUserID string
	router, _, _ := buildTestRouter(t, authHandler, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserID = r.Header.Get("X-Auth-UserId")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("X-Auth-UserId", "attacker_user") // spoofed
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if receivedUserID != "real_user" {
		t.Fatalf("spoofed X-Auth-UserId was not replaced: api-service got %q", receivedUserID)
	}
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, wantCode string) {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error.Code != wantCode {
		t.Fatalf("error code mismatch: got %q want %q", body.Error.Code, wantCode)
	}
}
