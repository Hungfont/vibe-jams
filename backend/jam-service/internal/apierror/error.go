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
	// CodeInternalError is returned on unexpected server failures.
	CodeInternalError = "internal_error"
)

// ErrorBody defines a stable error response contract for HTTP APIs.
type ErrorBody struct {
	Error ErrorPayload `json:"error"`
}

// ErrorPayload holds machine and human readable error details.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Write writes a standardized JSON error payload to response writer.
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
