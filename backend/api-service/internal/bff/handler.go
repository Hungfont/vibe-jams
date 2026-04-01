package bff

import (
	"encoding/json"
	"net/http"
	"strings"
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

	var req OrchestrateRequest
	if r.Body != nil {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
			writeJSON(w, http.StatusBadRequest, Envelope{Success: false, Error: &ErrorBody{Code: "invalid_input", Message: "invalid JSON body"}})
			return
		}
	}

	data, errBody, status := h.service.Orchestrate(r.Context(), sessionID, strings.TrimSpace(r.Header.Get("Authorization")), req)
	if errBody != nil {
		writeJSON(w, status, Envelope{Success: false, Error: errBody})
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Success: true, Data: data})
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
