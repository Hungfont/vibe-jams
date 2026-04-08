package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	sharedauth "video-streaming/backend/shared/auth"
)

const (
	testKID    = "test-kid"
	testSecret = "test-secret-at-least-32-chars-long"
)

func testVerifier(t *testing.T) *sharedauth.TokenVerifier {
	t.Helper()
	v, err := sharedauth.NewTokenVerifier(
		sharedauth.VerifierKey{KeyID: testKID, Secret: testSecret},
		nil,
	)
	if err != nil {
		t.Fatalf("build verifier: %v", err)
	}
	return v
}

func signToken(t *testing.T, kid, secret string, mc jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mc)
	token.Header["kid"] = kid
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	return signed
}

func validClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"userId":       "user_1",
		"plan":         "premium",
		"sessionState": "valid",
		"scope":        []any{"jam:read", "jam:control"},
		"sid":          "sess-1",
		"kid":          testKID,
		"exp":          float64(time.Now().Add(time.Hour).Unix()),
		"iat":          float64(time.Now().Unix()),
	}
}

func newTestMiddleware(t *testing.T) *authnMiddleware {
	t.Helper()
	return newAuthnMiddleware(testVerifier(t))
}

func TestAuthnMiddleware_MissingToken(t *testing.T) {
	t.Parallel()

	m := newTestMiddleware(t)
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

func TestAuthnMiddleware_InvalidToken(t *testing.T) {
	t.Parallel()

	m := newTestMiddleware(t)
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

func TestAuthnMiddleware_ExpiredToken(t *testing.T) {
	t.Parallel()

	m := newTestMiddleware(t)
	mc := validClaims()
	mc["exp"] = float64(time.Now().Add(-time.Hour).Unix())
	token := signToken(t, testKID, testSecret, mc)

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	ok := m.apply(rec, req)
	if ok {
		t.Fatal("expected middleware to reject expired token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorCode(t, rec, "invalid_token")
}

func TestAuthnMiddleware_NonValidSessionState(t *testing.T) {
	t.Parallel()

	m := newTestMiddleware(t)
	mc := validClaims()
	mc["sessionState"] = "invalid"
	token := signToken(t, testKID, testSecret, mc)

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
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

	m := newTestMiddleware(t)
	token := signToken(t, testKID, testSecret, validClaims())

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	ok := m.apply(rec, req)
	if !ok {
		t.Fatalf("expected middleware to pass request; response: %d", rec.Code)
	}
	if req.Header.Get("Authorization") != "Bearer "+token {
		t.Fatalf("Authorization header should be preserved after successful validation")
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

func TestAuthnMiddleware_CookieFallback_InjectsAuthorizationAndHeaders(t *testing.T) {
	t.Parallel()

	m := newTestMiddleware(t)
	token := signToken(t, testKID, testSecret, func() jwt.MapClaims {
		mc := validClaims()
		mc["userId"] = "user_cookie"
		return mc
	}())

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Cookie", "auth_token="+token)
	rec := httptest.NewRecorder()

	ok := m.apply(rec, req)
	if !ok {
		t.Fatalf("expected middleware to pass request; response: %d", rec.Code)
	}
	if req.Header.Get("Authorization") != "Bearer "+token {
		t.Fatalf("Authorization header should be set from cookie fallback: got %q", req.Header.Get("Authorization"))
	}
	if req.Header.Get("X-Auth-UserId") != "user_cookie" {
		t.Fatalf("X-Auth-UserId mismatch: got %q", req.Header.Get("X-Auth-UserId"))
	}
}

func TestAuthnMiddleware_AuthorizationHeaderPrecedenceOverCookie(t *testing.T) {
	t.Parallel()

	m := newTestMiddleware(t)
	headerToken := signToken(t, testKID, testSecret, func() jwt.MapClaims {
		mc := validClaims()
		mc["userId"] = "user_header"
		return mc
	}())
	cookieToken := signToken(t, testKID, testSecret, func() jwt.MapClaims {
		mc := validClaims()
		mc["userId"] = "user_cookie"
		return mc
	}())

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer "+headerToken)
	req.Header.Set("Cookie", "auth_token="+cookieToken)
	rec := httptest.NewRecorder()

	ok := m.apply(rec, req)
	if !ok {
		t.Fatalf("expected middleware to pass request; response: %d", rec.Code)
	}
	if req.Header.Get("X-Auth-UserId") != "user_header" {
		t.Fatalf("should use Authorization header over cookie: got userId=%q", req.Header.Get("X-Auth-UserId"))
	}
}

func TestAuthnMiddleware_ClientSpoofedXAuthHeaderStripped(t *testing.T) {
	t.Parallel()

	m := newTestMiddleware(t)
	token := signToken(t, testKID, testSecret, func() jwt.MapClaims {
		mc := validClaims()
		mc["userId"] = "real_user"
		mc["plan"] = "free"
		return mc
	}())

	req := httptest.NewRequest(http.MethodPost, "/v1/bff/mvp/sessions/jam_1/orchestration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
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

	m := newTestMiddleware(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	rec := httptest.NewRecorder()

	ok := m.apply(rec, req)
	if !ok {
		t.Fatal("expected public route to be passed through")
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
