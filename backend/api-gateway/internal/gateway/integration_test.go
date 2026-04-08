package gateway_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"video-streaming/backend/api-gateway/internal/config"
	"video-streaming/backend/api-gateway/internal/gateway"
)

const (
	integrationKID    = "integ-kid"
	integrationSecret = "integ-secret-at-least-32-chars!!"
)

func integrationSignToken(t *testing.T, mc jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mc)
	token.Header["kid"] = integrationKID
	signed, err := token.SignedString([]byte(integrationSecret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}

func integrationValidClaims(userID, plan, state string, scope []any) jwt.MapClaims {
	mc := jwt.MapClaims{
		"userId":       userID,
		"plan":         plan,
		"sessionState": state,
		"sid":          "sess-1",
		"kid":          integrationKID,
		"exp":          float64(time.Now().Add(time.Hour).Unix()),
		"iat":          float64(time.Now().Unix()),
	}
	if scope != nil {
		mc["scope"] = scope
	}
	return mc
}

func buildTestRouter(t *testing.T, authHandler, apiHandler http.HandlerFunc) (http.Handler, *httptest.Server, *httptest.Server) {
	t.Helper()
	authServer := httptest.NewServer(authHandler)
	apiServer := httptest.NewServer(apiHandler)
	t.Cleanup(func() {
		authServer.Close()
		apiServer.Close()
	})
	cfg := config.Config{
		ServerPort:         8085,
		AuthServiceURL:     authServer.URL,
		APIServiceURL:      apiServer.URL,
		AuthTimeout:        500_000_000,
		UpstreamTimeout:    2_000_000_000,
		JWTActiveKeyID:     integrationKID,
		JWTActiveKeySecret: integrationSecret,
	}
	router, err := gateway.NewRouter(cfg)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return router, authServer, apiServer
}

func TestIntegration_ValidToken_ProxiesToAPIService(t *testing.T) {
	t.Parallel()

	apiCalled := false
	var receivedUserID string
	var receivedAuthHeader string
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		receivedUserID = r.Header.Get("X-Auth-UserId")
		receivedAuthHeader = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"success":true,"data":{}}`))
	})

	token := integrationSignToken(t, integrationValidClaims("user_1", "premium", "valid", []any{"jam:read"}))
	router, _, _ := buildTestRouter(t,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }),
		apiHandler,
	)

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !apiCalled {
		t.Fatal("expected api-service to be called")
	}
	if receivedUserID != "user_1" {
		t.Fatalf("X-Auth-UserId not forwarded: got %q", receivedUserID)
	}
	if receivedAuthHeader != "Bearer "+token {
		t.Fatalf("Authorization header should be preserved for downstream compatibility: got %q", receivedAuthHeader)
	}
}

func TestIntegration_CookieFallback_ProxiesToAPIService(t *testing.T) {
	t.Parallel()

	apiCalled := false
	var receivedAuthHeader string
	token := integrationSignToken(t, integrationValidClaims("user_cookie", "premium", "valid", nil))
	router, _, _ := buildTestRouter(t,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiCalled = true
			receivedAuthHeader = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/bff/mvp/realtime/ws-config?sessionId=jam_1", nil)
	req.Header.Set("Cookie", "auth_token="+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}
	if !apiCalled {
		t.Fatal("expected api-service to be called")
	}
	if receivedAuthHeader != "Bearer "+token {
		t.Fatalf("api-service should receive Authorization derived from cookie fallback: got %q", receivedAuthHeader)
	}
}

func TestIntegration_InvalidToken_GatewayRejects401(t *testing.T) {
	t.Parallel()

	apiCalled := false
	router, _, _ := buildTestRouter(t,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }),
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			apiCalled = true
			w.WriteHeader(http.StatusOK)
		}),
	)

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

	token := integrationSignToken(t, integrationValidClaims("user_1", "premium", "invalid", nil))
	router, _, _ := buildTestRouter(t,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }),
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
	)

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorCode(t, rec, "invalid_token")
}

func TestIntegration_PublicAuthRoute_BypassesValidation(t *testing.T) {
	t.Parallel()

	loginCalled := false
	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

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

	token := integrationSignToken(t, integrationValidClaims("real_user", "free", "valid", nil))

	var receivedUserID string
	router, _, _ := buildTestRouter(t,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedUserID = r.Header.Get("X-Auth-UserId")
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Auth-UserId", "attacker_user")
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
