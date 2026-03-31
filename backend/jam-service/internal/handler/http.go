package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"video-streaming/backend/jams/internal/apierror"
	jamauth "video-streaming/backend/jams/internal/auth"
	"video-streaming/backend/jams/internal/model"
	"video-streaming/backend/jams/internal/service"
	sharedauth "video-streaming/backend/shared/auth"
)

// HTTPHandler exposes jam queue command/read endpoints over HTTP.
type HTTPHandler struct {
	service   *service.Service
	validator jamauth.Validator
}

// NewHTTPHandler creates a new HTTP handler for jam queue endpoints.
func NewHTTPHandler(service *service.Service, validator jamauth.Validator) *HTTPHandler {
	return &HTTPHandler{
		service:   service,
		validator: validator,
	}
}

// ServeHTTP dispatches all jam endpoints.
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/api/v1/jams/create":
		h.handleCreate(w, r)
		return
	case strings.HasPrefix(r.URL.Path, "/api/v1/jams/"):
		jamID, action, ok := parseJamSessionActionRoute(r.URL.Path)
		if !ok {
			break
		}
		switch action {
		case "join":
			h.handleJoin(jamID, w, r)
			return
		case "leave":
			h.handleLeave(jamID, w, r)
			return
		case "end":
			h.handleEnd(jamID, w, r)
			return
		}
	}

	jamID, action, ok := parseJamQueueRoute(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch action {
	case "add":
		h.handleAdd(jamID, w, r)
	case "remove":
		h.handleRemove(jamID, w, r)
	case "reorder":
		h.handleReorder(jamID, w, r)
	case "snapshot":
		h.handleSnapshot(jamID, w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleCreate starts a jam session for premium users.
func (h *HTTPHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := h.authorize(r.Context(), w, r, true)
	if !ok {
		return
	}
	session, err := h.service.CreateSession(claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

// handleJoin adds caller as participant in one active session.
func (h *HTTPHandler) handleJoin(jamID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := h.authorize(r.Context(), w, r, false)
	if !ok {
		return
	}
	session, err := h.service.JoinSession(jamID, claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// handleLeave removes caller from participant list.
// Host leave automatically ends the session.
func (h *HTTPHandler) handleLeave(jamID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := h.authorize(r.Context(), w, r, false)
	if !ok {
		return
	}
	session, err := h.service.LeaveSession(jamID, claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// handleEnd ends a jam session for premium users.
func (h *HTTPHandler) handleEnd(jamID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := h.authorize(r.Context(), w, r, true)
	if !ok {
		return
	}
	session, err := h.service.EndSession(jamID, claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// handleAdd processes queue-add command with idempotency behavior.
func (h *HTTPHandler) handleAdd(jamID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.AddQueueItemRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	snapshot, replayed, err := h.service.Add(jamID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	if replayed {
		w.Header().Set("X-Idempotent-Replay", "true")
	}
	writeJSON(w, http.StatusOK, snapshot)
}

// handleRemove processes queue-remove command.
func (h *HTTPHandler) handleRemove(jamID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.RemoveQueueItemRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	snapshot, err := h.service.Remove(jamID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, snapshot)
}

// handleReorder processes queue-reorder command with optimistic concurrency.
func (h *HTTPHandler) handleReorder(jamID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.ReorderQueueRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	snapshot, err := h.service.Reorder(jamID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, snapshot)
}

// handleSnapshot returns latest queue read model.
func (h *HTTPHandler) handleSnapshot(jamID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snapshot, err := h.service.Snapshot(jamID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, snapshot)
}

// decodeJSON parses JSON body and writes a consistent 400 response on errors.
func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		apierror.Write(w, http.StatusBadRequest, apierror.CodeInvalidInput, "invalid request body")
		return false
	}

	return true
}

// writeJSON marshals and writes successful JSON responses.
func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		apierror.Write(w, http.StatusInternalServerError, apierror.CodeInternalError, "failed to encode response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}

// writeServiceError maps domain/service failures into stable API error responses.
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case service.IsInvalidRequest(err):
		apierror.Write(w, http.StatusBadRequest, apierror.CodeInvalidInput, err.Error())
	case service.IsVersionConflict(err):
		apierror.Write(w, http.StatusConflict, apierror.CodeVersionConflict, err.Error())
	case service.IsNotFound(err):
		apierror.Write(w, http.StatusNotFound, apierror.CodeNotFound, err.Error())
	case service.IsHostOnly(err):
		apierror.Write(w, http.StatusForbidden, apierror.CodeHostOnly, "host only command")
	case service.IsSessionEnded(err):
		apierror.Write(w, http.StatusConflict, apierror.CodeSessionEnded, "session has ended")
	default:
		apierror.Write(w, http.StatusInternalServerError, apierror.CodeInternalError, "internal server error")
	}
}

// authorize validates token/session and checks optional premium entitlement.
func (h *HTTPHandler) authorize(ctx context.Context, w http.ResponseWriter, r *http.Request, requirePremium bool) (sharedauth.Claims, bool) {
	if h.validator == nil {
		apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "unauthorized")
		return sharedauth.Claims{}, false
	}
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "unauthorized")
		return sharedauth.Claims{}, false
	}

	claims, err := h.validator.ValidateBearerToken(ctx, authHeader)
	if err != nil {
		if errors.Is(err, jamauth.ErrUnauthorized) {
			apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "unauthorized")
			return sharedauth.Claims{}, false
		}
		apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "unauthorized")
		return sharedauth.Claims{}, false
	}
	if err := sharedauth.ValidateClaims(claims); err != nil {
		apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "unauthorized")
		return sharedauth.Claims{}, false
	}
	if strings.ToLower(claims.SessionState) != sharedauth.SessionStateValid {
		apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "unauthorized")
		return sharedauth.Claims{}, false
	}
	if requirePremium && !sharedauth.IsPremiumPlan(claims.Plan) {
		apierror.Write(w, http.StatusForbidden, apierror.CodePremiumRequired, "premium plan required")
		return sharedauth.Claims{}, false
	}

	return claims, true
}

// parseJamQueueRoute extracts jam ID and command action from queue endpoint path.
func parseJamQueueRoute(path string) (jamID string, action string, ok bool) {
	const prefix = "/api/v1/jams/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}

	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.Split(trimmed, "/")
	if len(parts) != 3 || parts[1] != "queue" {
		return "", "", false
	}
	if parts[0] == "" || parts[2] == "" {
		return "", "", false
	}

	return parts[0], parts[2], true
}

// parseJamSessionActionRoute extracts jam ID and lifecycle action from /api/v1/jams/{jamId}/{action} path.
func parseJamSessionActionRoute(path string) (jamID string, action string, ok bool) {
	const prefix = "/api/v1/jams/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}

	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 {
		return "", "", false
	}
	if parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	switch parts[1] {
	case "join", "leave", "end":
		return parts[0], parts[1], true
	default:
		return "", "", false
	}
}
