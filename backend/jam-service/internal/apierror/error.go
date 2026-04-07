package apierror

import (
	"encoding/json"
	"net/http"
)

const (
	// CodeInvalidInput is returned when request validation fails.
	CodeInvalidInput = "invalid_input"
	// CodeUnauthorized is returned for invalid or missing auth context.
	CodeUnauthorized = "unauthorized"
	// CodePremiumRequired is returned when premium entitlement is required.
	CodePremiumRequired = "premium_required"
	// CodeHostOnly is returned when a non-host executes host-only operation.
	CodeHostOnly = "host_only"
	// CodeNotFound is returned when a requested queue item cannot be found.
	CodeNotFound = "not_found"
	// CodeVersionConflict is returned when optimistic concurrency check fails.
	CodeVersionConflict = "version_conflict"
	// CodeSessionEnded is returned when write command targets ended session.
	CodeSessionEnded = "session_ended"
	// CodeTrackNotFound is returned when referenced track does not exist.
	CodeTrackNotFound = "track_not_found"
	// CodeTrackUnavailable is returned when referenced track is unavailable.
	CodeTrackUnavailable = "track_unavailable"
	// CodeModerationBlocked is returned when actor is blocked by moderation policy.
	CodeModerationBlocked = "moderation_blocked"
	// CodeInternalError is returned on unexpected server failures.
	CodeInternalError = "internal_error"
)

// ErrorBody defines a stable error response contract for HTTP APIs.
type ErrorBody struct {
	Error ErrorPayload `json:"error"`
}

// ErrorPayload holds machine and human readable error details.
type ErrorPayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Retry   *RetryGuidance `json:"retry,omitempty"`
}

// RetryGuidance carries authoritative metadata for deterministic conflict recovery.
type RetryGuidance struct {
	CurrentQueueVersion int64 `json:"currentQueueVersion"`
	PlaybackEpoch       int64 `json:"playbackEpoch,omitempty"`
}

// Write writes a standardized JSON error payload to response writer.
func Write(w http.ResponseWriter, statusCode int, code string, message string) {
	write(w, statusCode, code, message, nil)
}

// WriteWithRetry writes a standardized JSON error payload with optional retry guidance.
func WriteWithRetry(w http.ResponseWriter, statusCode int, code string, message string, retry *RetryGuidance) {
	write(w, statusCode, code, message, retry)
}

func write(w http.ResponseWriter, statusCode int, code string, message string, retry *RetryGuidance) {
	body := ErrorBody{
		Error: ErrorPayload{
			Code:    code,
			Message: message,
			Retry:   retry,
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
