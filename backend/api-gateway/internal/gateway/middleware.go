package gateway

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	sharedauth "video-streaming/backend/shared/auth"
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

// authnMiddleware enforces Bearer token validation for non-public routes
// using local JWT verification (no HTTP call to auth-service).
// On success it injects X-Auth-* headers and preserves Authorization for downstream compatibility.
type authnMiddleware struct {
	verifier *sharedauth.TokenVerifier
}

func newAuthnMiddleware(verifier *sharedauth.TokenVerifier) *authnMiddleware {
	return &authnMiddleware{verifier: verifier}
}

// apply wraps next with authN enforcement. Returns false and writes the error
// response if the request should not proceed.
func (m *authnMiddleware) apply(w http.ResponseWriter, r *http.Request) bool {
	for k := range r.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-auth-") {
			r.Header.Del(k)
		}
	}

	if isPublicRoute(r.Method, r.URL.Path) {
		return true
	}

	authHeader, fromCookie := resolveAuthHeader(r)
	if authHeader == "" {
		writeError(w, http.StatusUnauthorized, "missing_credentials", "missing authorization")
		return false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimSpace(token)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing_credentials", "missing authorization")
		return false
	}

	claims, err := m.verifier.VerifyAndExtractClaims(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return false
	}

	if strings.ToLower(strings.TrimSpace(claims.SessionState)) != "valid" {
		writeError(w, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return false
	}

	if fromCookie {
		r.Header.Set("Authorization", authHeader)
	}
	r.Header.Set(sharedauth.HeaderUserID, claims.UserID)
	r.Header.Set(sharedauth.HeaderPlan, claims.Plan)
	r.Header.Set(sharedauth.HeaderSessionState, claims.SessionState)
	if len(claims.Scope) > 0 {
		r.Header.Set(sharedauth.HeaderScope, strings.Join(claims.Scope, ","))
	}

	return true
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
