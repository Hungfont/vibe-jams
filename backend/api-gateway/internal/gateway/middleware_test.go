package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestMiddleware(authServerURL string) *authnMiddleware {
	return newAuthnMiddleware(authServerURL, 500*time.Millisecond)
}

func applyAndRecord(t *testing.T, m *authnMiddleware, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	m.apply(rec, r)
	return rec
}

func TestAuthnMiddleware_MissingToken(t *testing.T) {
	t.Parallel()

	m := newTestMiddleware("http://127.0.0.1:0")
	req := httptest.NewRequest(http.MethodGet, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	rec := httptest.NewRecorder()

	ok := m.apply(rec, req)
	if ok {
		t.Fatal("expected middleware to reject request")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorCode(t, rec, "missing_credentials")
}

func TestAuthnMiddleware_InvalidToken_AuthReturns401(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"code":"invalid_token","message":"invalid"}}`))
	}))
	defer ts.Close()

	m := newTestMiddleware(ts.URL)
	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()

	ok := m.apply(rec, req)
	if ok {
		t.Fatal("expected middleware to reject request")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorCode(t, rec, "invalid_token")
}

func TestAuthnMiddleware_AuthServiceTimeout(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hang until client context is cancelled.
		<-r.Context().Done()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m := newAuthnMiddleware(ts.URL, 20*time.Millisecond)
	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()

	ok := m.apply(rec, req)
	if ok {
		t.Fatal("expected middleware to reject request")
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusServiceUnavailable)
	}
	assertErrorCode(t, rec, "auth_service_unavailable")
}

func TestAuthnMiddleware_NonValidSessionState(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"userId":"user_1","plan":"premium","sessionState":"invalid"}`))
	}))
	defer ts.Close()

	m := newTestMiddleware(ts.URL)
	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer revoked-token")
	rec := httptest.NewRecorder()

	ok := m.apply(rec, req)
	if ok {
		t.Fatal("expected middleware to reject revoked session")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorCode(t, rec, "invalid_token")
}

func TestAuthnMiddleware_ValidToken_InjectsHeaders(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"userId":"user_1","plan":"premium","sessionState":"valid","scope":["jam:read","jam:control"]}`))
	}))
	defer ts.Close()

	m := newTestMiddleware(ts.URL)
	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	ok := m.apply(rec, req)
	if !ok {
		t.Fatalf("expected middleware to pass request; response: %d", rec.Code)
	}
	if req.Header.Get("Authorization") != "" {
		t.Fatal("Authorization header should be stripped after successful validation")
	}
	if req.Header.Get("X-Auth-UserId") != "user_1" {
		t.Fatalf("X-Auth-UserId mismatch: got %q", req.Header.Get("X-Auth-UserId"))
	}
	if req.Header.Get("X-Auth-Plan") != "premium" {
		t.Fatalf("X-Auth-Plan mismatch: got %q", req.Header.Get("X-Auth-Plan"))
	}
	if req.Header.Get("X-Auth-SessionState") != "valid" {
		t.Fatalf("X-Auth-SessionState mismatch: got %q", req.Header.Get("X-Auth-SessionState"))
	}
	if req.Header.Get("X-Auth-Scope") != "jam:read,jam:control" {
		t.Fatalf("X-Auth-Scope mismatch: got %q", req.Header.Get("X-Auth-Scope"))
	}
}

func TestAuthnMiddleware_ClientSpoofedXAuthHeaderStripped(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"userId":"real_user","plan":"free","sessionState":"valid"}`))
	}))
	defer ts.Close()

	m := newTestMiddleware(ts.URL)
	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	// Attacker-supplied header that should be stripped.
	req.Header.Set("X-Auth-UserId", "attacker_user")
	rec := httptest.NewRecorder()

	ok := m.apply(rec, req)
	if !ok {
		t.Fatalf("expected middleware to pass; response: %d", rec.Code)
	}
	if req.Header.Get("X-Auth-UserId") != "real_user" {
		t.Fatalf("spoofed X-Auth-UserId was not replaced: got %q", req.Header.Get("X-Auth-UserId"))
	}
}

func TestAuthnMiddleware_PublicRouteBypass(t *testing.T) {
	t.Parallel()

	// Auth server should NOT be called for public routes.
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m := newTestMiddleware(ts.URL)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	// No Authorization header — public route should pass through.
	rec := httptest.NewRecorder()

	ok := m.apply(rec, req)
	if !ok {
		t.Fatal("expected public route to be passed through")
	}
	if called {
		t.Fatal("auth-service validate should not be called for public routes")
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
