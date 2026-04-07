package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// publicRoutes are forwarded to auth-service without token verification.
var publicRoutes = []struct {
	method string
	path   string
}{
	{http.MethodPost, "/v1/auth/login"},
	{http.MethodPost, "/v1/auth/refresh"},
	{http.MethodPost, "/v1/auth/logout"},
	{http.MethodGet, "/v1/auth/me"},
}

// validateClaims is the response shape from POST /internal/v1/auth/validate.
type validateClaims struct {
	UserID       string   `json:"userId"`
	Plan         string   `json:"plan"`
	SessionState string   `json:"sessionState"`
	Scope        []string `json:"scope,omitempty"`
}

// authnMiddleware enforces Bearer token validation for non-public routes.
// On success it injects X-Auth-* headers and preserves Authorization for downstream compatibility.
type authnMiddleware struct {
	authBaseURL string
	authClient  *http.Client
}

func newAuthnMiddleware(authBaseURL string, timeout time.Duration) *authnMiddleware {
	return &authnMiddleware{
		authBaseURL: strings.TrimRight(authBaseURL, "/"),
		authClient:  &http.Client{Timeout: timeout},
	}
}

// apply wraps next with authN enforcement. Returns false and writes the error
// response if the request should not proceed.
func (m *authnMiddleware) apply(w http.ResponseWriter, r *http.Request) bool {
	// Always strip client-supplied X-Auth-* headers to prevent spoofing.
	for k := range r.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-auth-") {
			r.Header.Del(k)
		}
	}

	// Public routes bypass token verification.
	if isPublicRoute(r.Method, r.URL.Path) {
		return true
	}

	authHeader, fromCookie := resolveAuthHeader(r)
	if authHeader == "" {
		writeError(w, http.StatusUnauthorized, "missing_credentials", "missing authorization")
		return false
	}

	claims, err := m.validate(r.Context(), authHeader)
	if err != nil {
		if isAuthUnavailable(err) {
			writeError(w, http.StatusServiceUnavailable, "auth_service_unavailable", "auth service unavailable")
		} else {
			writeError(w, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		}
		return false
	}

	if strings.ToLower(strings.TrimSpace(claims.SessionState)) != "valid" {
		writeError(w, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return false
	}

	// Preserve Authorization for downstream services that still validate bearer tokens,
	// and inject verified claim headers for services that trust gateway identity context.
	if fromCookie {
		r.Header.Set("Authorization", authHeader)
	}
	r.Header.Set("X-Auth-UserId", claims.UserID)
	r.Header.Set("X-Auth-Plan", claims.Plan)
	r.Header.Set("X-Auth-SessionState", claims.SessionState)
	if len(claims.Scope) > 0 {
		r.Header.Set("X-Auth-Scope", strings.Join(claims.Scope, ","))
	}

	return true
}

func (m *authnMiddleware) validate(ctx context.Context, authHeader string) (validateClaims, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.authBaseURL+"/internal/v1/auth/validate", bytes.NewReader(nil))
	if err != nil {
		return validateClaims{}, fmt.Errorf("build validate request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := m.authClient.Do(req)
	if err != nil {
		return validateClaims{}, errAuthUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode >= http.StatusInternalServerError {
			return validateClaims{}, errAuthUnavailable
		}
		return validateClaims{}, errInvalidToken
	}

	var claims validateClaims
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return validateClaims{}, errInvalidToken
	}
	if strings.TrimSpace(claims.UserID) == "" {
		return validateClaims{}, errInvalidToken
	}
	return claims, nil
}

var (
	errAuthUnavailable = fmt.Errorf("auth service unavailable")
	errInvalidToken    = fmt.Errorf("invalid token")
)

func isAuthUnavailable(err error) bool {
	return err == errAuthUnavailable
}

func isPublicRoute(method, path string) bool {
	for _, pr := range publicRoutes {
		if pr.method == method && pr.path == path {
			return true
		}
	}
	return false
}

func resolveAuthHeader(r *http.Request) (header string, fromCookie bool) {
	header = strings.TrimSpace(r.Header.Get("Authorization"))
	if header != "" {
		return header, false
	}

	cookieToken := resolveAuthCookieToken(r)
	if cookieToken == "" {
		return "", false
	}

	return "Bearer " + cookieToken, true
}

func resolveAuthCookieToken(r *http.Request) string {
	for _, cookieName := range []string{"auth_token", "token"} {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			continue
		}
		value := strings.TrimSpace(cookie.Value)
		if value == "" {
			continue
		}

		decoded, err := url.QueryUnescape(value)
		if err == nil {
			decoded = strings.TrimSpace(decoded)
			if decoded != "" {
				return decoded
			}
		}

		return value
	}

	return ""
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	body, _ := json.Marshal(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
