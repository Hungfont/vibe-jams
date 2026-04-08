package httpserver

import (
	"bufio"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"video-streaming/backend/auth-service/internal/auth"
)

func TestValidateToken(t *testing.T) {
	t.Parallel()

	handler := NewHandler(newTestAuthenticator(t, testAuthOptions{}))
	router := handler.Router()

	tests := []struct {
		name           string
		authHeader     string
		wantStatusCode int
		wantErrorCode  string
		wantUserID     string
	}{
		{
			name:           "valid premium token",
			authHeader:     "Bearer token-premium-valid",
			wantStatusCode: http.StatusOK,
			wantUserID:     "user-premium-1",
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer unknown-token",
			wantStatusCode: http.StatusUnauthorized,
			wantErrorCode:  "unauthorized",
		},
		{
			name:           "revoked session token returns claims contract",
			authHeader:     "Bearer token-premium-revoked",
			wantStatusCode: http.StatusOK,
			wantUserID:     "user-premium-1",
		},
		{
			name:           "missing token",
			authHeader:     "",
			wantStatusCode: http.StatusUnauthorized,
			wantErrorCode:  "unauthorized",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/internal/v1/auth/validate", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatusCode {
				t.Fatalf("status mismatch: got %d want %d", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK {
				var body struct {
					UserID string `json:"userId"`
				}
				if err := json.NewDecoder(strings.NewReader(rec.Body.String())).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.UserID != tt.wantUserID {
					t.Fatalf("user mismatch: got %q want %q", body.UserID, tt.wantUserID)
				}
				return
			}

			var body struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.NewDecoder(strings.NewReader(rec.Body.String())).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Error.Code != tt.wantErrorCode {
				t.Fatalf("error code mismatch: got %q want %q", body.Error.Code, tt.wantErrorCode)
			}
		})
	}
}

func TestPublicAuthFlowEndpoints(t *testing.T) {
	t.Parallel()

	handler := NewHandler(newTestAuthenticator(t, testAuthOptions{}))
	router := handler.Router()

	loginRec := httptest.NewRecorder()
	loginReq := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{"identity":"premium@example.com","password":"premium-pass"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status mismatch: got %d want %d", loginRec.Code, http.StatusOK)
	}

	var loginBody struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		Claims       struct {
			UserID string   `json:"userId"`
			Scope  []string `json:"scope"`
		} `json:"claims"`
	}
	if err := json.NewDecoder(strings.NewReader(loginRec.Body.String())).Decode(&loginBody); err != nil {
		t.Fatalf("decode login body: %v", err)
	}
	if loginBody.AccessToken == "" || loginBody.RefreshToken == "" {
		t.Fatal("expected tokens in login response")
	}
	if loginBody.Claims.UserID != "user-premium-1" {
		t.Fatalf("unexpected login userId: %s", loginBody.Claims.UserID)
	}
	if len(loginBody.Claims.Scope) == 0 {
		t.Fatal("expected scope in login claims")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginBody.AccessToken)
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me status mismatch: got %d want %d", meRec.Code, http.StatusOK)
	}

	refreshBody := `{"refreshToken":"` + loginBody.RefreshToken + `"}`
	refreshReq := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", strings.NewReader(refreshBody))
	refreshReq.Header.Set("Content-Type", "application/json")
	refreshRec := httptest.NewRecorder()
	router.ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status mismatch: got %d want %d", refreshRec.Code, http.StatusOK)
	}

	var refreshResp struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(strings.NewReader(refreshRec.Body.String())).Decode(&refreshResp); err != nil {
		t.Fatalf("decode refresh body: %v", err)
	}
	if refreshResp.RefreshToken == "" || refreshResp.RefreshToken == loginBody.RefreshToken {
		t.Fatal("expected rotated refresh token")
	}

	reuseReq := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", strings.NewReader(refreshBody))
	reuseReq.Header.Set("Content-Type", "application/json")
	reuseRec := httptest.NewRecorder()
	router.ServeHTTP(reuseRec, reuseReq)
	if reuseRec.Code != http.StatusUnauthorized {
		t.Fatalf("reuse status mismatch: got %d want %d", reuseRec.Code, http.StatusUnauthorized)
	}

	rotatedReq := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", strings.NewReader(`{"refreshToken":"`+refreshResp.RefreshToken+`"}`))
	rotatedReq.Header.Set("Content-Type", "application/json")
	rotatedRec := httptest.NewRecorder()
	router.ServeHTTP(rotatedRec, rotatedReq)
	if rotatedRec.Code != http.StatusUnauthorized {
		t.Fatalf("family revocation status mismatch: got %d want %d", rotatedRec.Code, http.StatusUnauthorized)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	t.Parallel()

	handler := NewHandler(newTestAuthenticator(t, testAuthOptions{}))
	router := handler.Router()

	loginRec := httptest.NewRecorder()
	loginReq := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{"identity":"free@example.com","password":"free-pass"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status mismatch: got %d", loginRec.Code)
	}

	var loginBody struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(strings.NewReader(loginRec.Body.String())).Decode(&loginBody); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", strings.NewReader(`{"refreshToken":"`+loginBody.RefreshToken+`"}`))
	logoutReq.Header.Set("Content-Type", "application/json")
	logoutRec := httptest.NewRecorder()
	router.ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusOK {
		t.Fatalf("logout status mismatch: got %d", logoutRec.Code)
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", strings.NewReader(`{"refreshToken":"`+loginBody.RefreshToken+`"}`))
	refreshReq.Header.Set("Content-Type", "application/json")
	refreshRec := httptest.NewRecorder()
	router.ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked refresh token unauthorized, got %d", refreshRec.Code)
	}
}

func TestLoginRateLimitAndLockout(t *testing.T) {
	t.Parallel()

	handler := NewHandler(newTestAuthenticator(t, testAuthOptions{
		LoginLimit:       2,
		LoginWindow:      time.Minute,
		LockoutThreshold: 2,
		LockoutBackoff:   time.Minute,
	}))
	router := handler.Router()

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{"identity":"premium@example.com","password":"wrong"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rec, req)
		if i == 0 && rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d expected unauthorized, got %d", i+1, rec.Code)
		}
		if i == 1 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("attempt %d expected lockout throttled, got %d", i+1, rec.Code)
		}
	}

	rateRec := httptest.NewRecorder()
	rateReq := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{"identity":"premium@example.com","password":"premium-pass"}`))
	rateReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rateRec, rateReq)
	if rateRec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected rate limit throttled, got %d", rateRec.Code)
	}
}

func TestRequestLoggingMiddleware_PreservesFlusherAndHijacker(t *testing.T) {
	t.Parallel()

	wrapped := requestLoggingMiddleware("auth-service", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected wrapped writer to implement http.Flusher")
		}
		flusher.Flush()

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("expected wrapped writer to implement http.Hijacker")
		}
		if _, _, err := hijacker.Hijack(); err != nil {
			t.Fatalf("expected hijack to delegate to underlying writer: %v", err)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	rec := newHijackableRecorder(t)
	wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusNoContent)
	}
	if rec.flushCount != 1 {
		t.Fatalf("flush count mismatch: got %d want 1", rec.flushCount)
	}
}

type testAuthOptions struct {
	LoginLimit       int
	LoginWindow      time.Duration
	RefreshLimit     int
	RefreshWindow    time.Duration
	LockoutThreshold int
	LockoutBackoff   time.Duration
}

func newTestAuthenticator(t *testing.T, opts testAuthOptions) auth.Authenticator {
	t.Helper()

	now := time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC)
	keyRing, err := auth.NewKeyRing(auth.SigningKey{KeyID: "kid-active", Secret: "test-secret"}, nil)
	if err != nil {
		t.Fatalf("new key ring: %v", err)
	}

	if opts.LoginLimit <= 0 {
		opts.LoginLimit = 50
	}
	if opts.LoginWindow <= 0 {
		opts.LoginWindow = time.Minute
	}
	if opts.RefreshLimit <= 0 {
		opts.RefreshLimit = 50
	}
	if opts.RefreshWindow <= 0 {
		opts.RefreshWindow = time.Minute
	}
	if opts.LockoutThreshold <= 0 {
		opts.LockoutThreshold = 5
	}
	if opts.LockoutBackoff <= 0 {
		opts.LockoutBackoff = time.Minute
	}

	svc, err := auth.NewService(auth.ServiceConfig{
		Credentials:     auth.NewInMemoryCredentialStore(),
		SessionStore:    auth.NewInMemorySessionStore(),
		KeyRing:         keyRing,
		FixtureClaims:   auth.DefaultFixtureClaims(),
		AuditLogger:     auth.NoopAuditLogger{},
		LoginLimiter:    auth.NewFixedWindowLimiter(opts.LoginLimit, opts.LoginWindow, func() time.Time { return now }),
		RefreshLimiter:  auth.NewFixedWindowLimiter(opts.RefreshLimit, opts.RefreshWindow, func() time.Time { return now }),
		LockoutTracker:  auth.NewLockoutTracker(opts.LockoutThreshold, opts.LockoutBackoff, func() time.Time { return now }),
		AccessTokenTTL:  10 * time.Minute,
		RefreshTokenTTL: 30 * time.Minute,
		Now:             func() time.Time { return now },
		RandomReader:    &incrementingReader{},
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	return svc
}

type incrementingReader struct {
	next byte
}

func (r *incrementingReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	for i := range p {
		r.next++
		p[i] = r.next
	}
	return len(p), nil
}

type hijackableRecorder struct {
	*httptest.ResponseRecorder
	conn       net.Conn
	rw         *bufio.ReadWriter
	flushCount int
}

func newHijackableRecorder(t *testing.T) *hijackableRecorder {
	t.Helper()

	client, server := net.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})

	return &hijackableRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		conn:             client,
		rw:               bufio.NewReadWriter(bufio.NewReader(server), bufio.NewWriter(server)),
	}
}

func (r *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.conn, r.rw, nil
}

func (r *hijackableRecorder) Flush() {
	r.flushCount++
}
