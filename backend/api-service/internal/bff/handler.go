package bff

import (
	"encoding/json"
	"net/http"
	"strings"

	sharedauth "video-streaming/backend/shared/auth"
)

// Handler exposes BFF orchestration endpoint routes.
type Handler struct {
	service *Service
}

// NewHandler builds a BFF HTTP handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ServeHTTP routes BFF orchestration requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := parseOrchestrationRoute(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, xAuthHeaders, ok := extractGatewayIdentity(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, Envelope{Success: false, Error: &ErrorBody{Code: "unauthorized", Message: "missing or invalid identity context"}})
		return
	}

	var req OrchestrateRequest
	if r.Body != nil {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
			writeJSON(w, http.StatusBadRequest, Envelope{Success: false, Error: &ErrorBody{Code: "invalid_input", Message: "invalid JSON body"}})
			return
		}
	}

	data, errBody, status := h.service.Orchestrate(r.Context(), sessionID, claims, xAuthHeaders, req)
	if errBody != nil {
		writeJSON(w, status, Envelope{Success: false, Error: errBody})
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Success: true, Data: data})
}

// extractGatewayIdentity reads X-Auth-* headers injected by api-gateway and validates them.
// Returns the normalized claims, the X-Auth header set to forward downstream, and whether the identity is valid.
func extractGatewayIdentity(r *http.Request) (sharedauth.Claims, http.Header, bool) {
	userID := strings.TrimSpace(r.Header.Get("X-Auth-UserId"))
	if userID == "" {
		return sharedauth.Claims{}, nil, false
	}
	plan := strings.TrimSpace(r.Header.Get("X-Auth-Plan"))
	sessionState := strings.TrimSpace(r.Header.Get("X-Auth-SessionState"))
	if strings.ToLower(sessionState) != sharedauth.SessionStateValid {
		return sharedauth.Claims{}, nil, false
	}

	var scope []string
	if raw := strings.TrimSpace(r.Header.Get("X-Auth-Scope")); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				scope = append(scope, trimmed)
			}
		}
	}

	claims := sharedauth.Claims{
		UserID:       userID,
		Plan:         plan,
		SessionState: sessionState,
		Scope:        scope,
	}

	xAuthHeaders := make(http.Header)
	for k, vs := range r.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-auth-") {
			xAuthHeaders[k] = vs
		}
	}

	return claims, xAuthHeaders, true
}

func parseOrchestrationRoute(path string) (string, bool) {
	const prefix = "/v1/bff/mvp/sessions/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 {
		return "", false
	}
	if parts[0] == "" || parts[1] != "orchestration" {
		return "", false
	}
	return parts[0], true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
