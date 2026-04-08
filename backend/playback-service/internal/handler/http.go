package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"video-streaming/backend/playback-service/internal/apierror"
	playbackauth "video-streaming/backend/playback-service/internal/auth"
	"video-streaming/backend/playback-service/internal/model"
	"video-streaming/backend/playback-service/internal/service"
	sharedauth "video-streaming/backend/shared/auth"
)

// CommandService exposes playback command execution behavior.
type CommandService interface {
	ExecuteCommand(ctx context.Context, sessionID string, actorUserID string, req model.PlaybackCommandRequest) (model.CommandAcceptedResponse, error)
}

// HTTPHandler serves playback command endpoint.
type HTTPHandler struct {
	service   CommandService
	validator playbackauth.Validator
}

// NewHTTPHandler builds playback HTTP handler.
func NewHTTPHandler(service CommandService, validator playbackauth.Validator) *HTTPHandler {
	return &HTTPHandler{
		service:   service,
		validator: validator,
	}
}

// ServeHTTP dispatches playback command endpoint by route.
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := parsePlaybackCommandRoute(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	h.handlePlaybackCommand(sessionID, w, r)
}

func (h *HTTPHandler) handlePlaybackCommand(sessionID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, authorized := h.authorize(r.Context(), w, r)
	if !authorized {
		return
	}

	var req model.PlaybackCommandRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	resp, err := h.service.ExecuteCommand(r.Context(), sessionID, claims.UserID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, resp)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		apierror.Write(w, http.StatusBadRequest, apierror.CodeInvalidInput, "invalid request body")
		return false
	}
	return true
}

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

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case service.IsInvalidRequest(err):
		apierror.Write(w, http.StatusBadRequest, apierror.CodeInvalidInput, err.Error())
	case service.IsTrackNotFound(err):
		apierror.Write(w, http.StatusNotFound, apierror.CodeTrackNotFound, "track not found")
	case service.IsTrackUnavailable(err):
		apierror.Write(w, http.StatusConflict, apierror.CodeTrackUnavailable, "track unavailable")
	case service.IsHostOnly(err):
		apierror.Write(w, http.StatusForbidden, apierror.CodeHostOnly, "host only command")
	case service.IsVersionConflict(err):
		if retry, ok := service.ConflictRetryFromError(err); ok {
			apierror.WriteWithRetry(w, http.StatusConflict, apierror.CodeVersionConflict, "stale queue version", &apierror.RetryGuidance{
				CurrentQueueVersion: retry.CurrentQueueVersion,
				PlaybackEpoch:       retry.PlaybackEpoch,
			})
			return
		}
		apierror.Write(w, http.StatusConflict, apierror.CodeVersionConflict, "stale queue version")
	case service.IsSessionEnded(err):
		apierror.Write(w, http.StatusConflict, apierror.CodeSessionEnded, "session has ended")
	case service.IsNotFound(err):
		apierror.Write(w, http.StatusNotFound, apierror.CodeNotFound, "session not found")
	default:
		apierror.Write(w, http.StatusInternalServerError, apierror.CodeInternalError, "internal server error")
	}
}

// authorize validates token/session. It first tries gateway-injected X-Auth-*
// headers (zero-cost path), then falls back to the configured Validator.
func (h *HTTPHandler) authorize(ctx context.Context, w http.ResponseWriter, r *http.Request) (sharedauth.Claims, bool) {
	claims, ok := sharedauth.ExtractClaimsFromHeaders(r.Header)
	if !ok {
		if h.validator == nil {
			apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "unauthorized")
			return sharedauth.Claims{}, false
		}
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "unauthorized")
			return sharedauth.Claims{}, false
		}

		var err error
		claims, err = h.validator.ValidateBearerToken(ctx, authHeader)
		if err != nil {
			apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "unauthorized")
			return sharedauth.Claims{}, false
		}
		if err := sharedauth.ValidateClaims(claims); err != nil {
			apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "unauthorized")
			return sharedauth.Claims{}, false
		}
	}

	if strings.ToLower(claims.SessionState) != sharedauth.SessionStateValid {
		apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "unauthorized")
		return sharedauth.Claims{}, false
	}
	return claims, true
}

func parsePlaybackCommandRoute(path string) (sessionID string, ok bool) {
	const prefix = "/v1/jam/sessions/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.Split(trimmed, "/")
	if len(parts) != 3 || parts[1] != "playback" || parts[2] != "commands" {
		return "", false
	}
	if strings.TrimSpace(parts[0]) == "" {
		return "", false
	}
	return parts[0], true
}
