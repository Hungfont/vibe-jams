package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"video-streaming/backend/auth-service/internal/auth"
	sharedauth "video-streaming/backend/shared/auth"
)

const unauthorizedCode = "unauthorized"

// Handler serves auth validation endpoints consumed by backend services.
type Handler struct {
	authenticator auth.Authenticator
}

// NewHandler creates a new auth-service HTTP handler.
func NewHandler(authenticator auth.Authenticator) *Handler {
	return &Handler{authenticator: authenticator}
}

// Router builds HTTP routes for auth-service.
func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/login", h.login)
	mux.HandleFunc("/v1/auth/refresh", h.refresh)
	mux.HandleFunc("/v1/auth/logout", h.logout)
	mux.HandleFunc("/v1/auth/me", h.me)
	mux.HandleFunc("/internal/v1/auth/validate", h.validateToken)
	mux.HandleFunc("/healthz", h.healthz)
	return mux
}

// validateToken validates a bearer token and returns normalized claims.
func (h *Handler) validateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, ok := parseBearerToken(r.Header.Get("Authorization"))
	if !ok {
		writeError(w, http.StatusUnauthorized, unauthorizedCode, "missing or invalid bearer token")
		return
	}

	claims, err := h.authenticator.ValidateBearerToken(token)
	if err != nil {
		if errors.Is(err, auth.ErrTokenNotFound) {
			writeError(w, http.StatusUnauthorized, unauthorizedCode, "invalid or expired token")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	if err := sharedauth.ValidateClaims(claims); err != nil {
		writeError(w, http.StatusUnauthorized, unauthorizedCode, "unauthorized claims")
		return
	}

	writeJSON(w, http.StatusOK, claims)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Identity string `json:"identity"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid request payload")
		return
	}

	pair, err := h.authenticator.Login(r.Context(), auth.LoginRequest{
		Identity: payload.Identity,
		Password: payload.Password,
		IP:       clientIP(r),
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"accessToken":  pair.AccessToken,
		"refreshToken": pair.RefreshToken,
		"tokenType":    "Bearer",
		"expiresAt":    pair.AccessExpires.UTC().Format(time.RFC3339),
		"claims":       pair.Claims,
	})
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid request payload")
		return
	}

	pair, err := h.authenticator.Refresh(r.Context(), auth.RefreshRequest{
		RefreshToken: payload.RefreshToken,
		IP:           clientIP(r),
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"accessToken":  pair.AccessToken,
		"refreshToken": pair.RefreshToken,
		"tokenType":    "Bearer",
		"expiresAt":    pair.AccessExpires.UTC().Format(time.RFC3339),
		"claims":       pair.Claims,
	})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid request payload")
		return
	}

	if err := h.authenticator.Logout(r.Context(), auth.LogoutRequest{
		RefreshToken: payload.RefreshToken,
		IP:           clientIP(r),
	}); err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, ok := parseBearerToken(r.Header.Get("Authorization"))
	if !ok {
		writeError(w, http.StatusUnauthorized, unauthorizedCode, "missing or invalid bearer token")
		return
	}
	claims, err := h.authenticator.Me(r.Context(), token)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, claims)
}

// healthz reports service liveness.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "auth-service",
	})
}

// parseBearerToken extracts token from Authorization header.
func parseBearerToken(header string) (string, bool) {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 {
		return "", false
	}
	if !strings.EqualFold(parts[0], "bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

func decodeJSON(r *http.Request, out any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrRateLimited), errors.Is(err, auth.ErrLockedOut):
		writeError(w, http.StatusTooManyRequests, "throttled", "too many attempts")
	case errors.Is(err, auth.ErrInvalidCredentials), errors.Is(err, auth.ErrUnauthorized), errors.Is(err, auth.ErrTokenNotFound), errors.Is(err, auth.ErrRefreshReuseDetected):
		writeError(w, http.StatusUnauthorized, unauthorizedCode, "invalid or expired auth context")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	hostPort := strings.TrimSpace(r.RemoteAddr)
	if hostPort == "" {
		return "unknown"
	}
	if addr, err := netip.ParseAddrPort(hostPort); err == nil {
		return addr.Addr().String()
	}
	return hostPort
}

// writeJSON writes successful JSON payloads.
func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}

// writeError writes standardized error responses for auth-service.
func writeError(w http.ResponseWriter, statusCode int, code string, message string) {
	writeJSON(w, statusCode, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
