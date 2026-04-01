package apierror

import (
	"encoding/json"
	"net/http"
)

const (
	// CodeInvalidInput is returned when request validation fails.
	CodeInvalidInput = "invalid_input"
	// CodeUnauthorized is returned for missing/invalid authentication context.
	CodeUnauthorized = "unauthorized"
	// CodeHostOnly is returned when a non-host user executes host-only commands.
	CodeHostOnly = "host_only"
	// CodeVersionConflict is returned when optimistic concurrency check fails.
	CodeVersionConflict = "version_conflict"
	// CodeSessionEnded is returned when command targets ended session.
	CodeSessionEnded = "session_ended"
	// CodeTrackNotFound is returned when referenced track does not exist.
	CodeTrackNotFound = "track_not_found"
	// CodeTrackUnavailable is returned when referenced track is unavailable.
	CodeTrackUnavailable = "track_unavailable"
	// CodeNotFound is returned when requested session context is missing.
	CodeNotFound = "not_found"
	// CodeInternalError is returned on unexpected server failures.
	CodeInternalError = "internal_error"
)

// ErrorBody defines stable JSON error response contract.
type ErrorBody struct {
	Error ErrorPayload `json:"error"`
}

// ErrorPayload carries machine and human-readable details.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Write writes standardized JSON error payload.
func Write(w http.ResponseWriter, statusCode int, code string, message string) {
	body := ErrorBody{
		Error: ErrorPayload{
			Code:    code,
			Message: message,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(payload)
}
