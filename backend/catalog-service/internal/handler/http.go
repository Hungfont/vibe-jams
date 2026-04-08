package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	sharedcatalog "video-streaming/backend/shared/catalog"

	"video-streaming/backend/catalog-service/internal/repository"
)

const (
	codeTrackNotFound = "track_not_found"
)

// HTTPHandler serves catalog lookup endpoints.
type HTTPHandler struct {
	store repository.Store
}

// NewHTTPHandler builds a catalog HTTP handler.
func NewHTTPHandler(store repository.Store) *HTTPHandler {
	return &HTTPHandler{store: store}
}

// ServeHTTP dispatches internal catalog routes.
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	trackID, ok := parseTrackLookupRoute(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	h.handleLookup(trackID, w, r)
}

func (h *HTTPHandler) handleLookup(trackID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	track, ok := h.store.FindByID(trackID)
	if !ok {
		writeError(w, http.StatusNotFound, codeTrackNotFound, "track not found")
		return
	}

	writeJSON(w, http.StatusOK, sharedcatalog.LookupResponse{
		TrackID:      track.TrackID,
		IsPlayable:   track.IsPlayable,
		ReasonCode:   track.ReasonCode,
		PolicyStatus: track.PolicyStatus,
		PolicyReason: track.PolicyReason,
		Title:        track.Title,
		Artist:       track.Artist,
	})
}

func parseTrackLookupRoute(path string) (trackID string, ok bool) {
	const prefix = "/internal/v1/catalog/tracks/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	trackID = strings.TrimSpace(strings.TrimPrefix(path, prefix))
	if trackID == "" {
		return "", false
	}
	return trackID, true
}

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

func writeError(w http.ResponseWriter, statusCode int, code string, message string) {
	writeJSON(w, statusCode, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}
