package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"video-streaming/backend/auth-service/internal/auth"
	sharedauth "video-streaming/backend/shared/auth"
)

const unauthorizedCode = "unauthorized"

// Handler serves auth validation endpoints consumed by backend services.
type Handler struct {
	validator auth.Validator
}

// NewHandler creates a new auth-service HTTP handler.
func NewHandler(validator auth.Validator) *Handler {
	return &Handler{validator: validator}
}

// Router builds HTTP routes for auth-service.
func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
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

	claims, err := h.validator.ValidateBearerToken(token)
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
