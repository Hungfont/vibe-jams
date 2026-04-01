package bff

import (
	"context"
	"errors"
	"fmt"
	"net"
)

var (
	// ErrUnauthorized indicates authentication/claim validation failure.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrDependencyUnavailable indicates dependency service unavailable.
	ErrDependencyUnavailable = errors.New("dependency unavailable")
	// ErrDependencyTimeout indicates dependency call timed out.
	ErrDependencyTimeout = errors.New("dependency timeout")
	// ErrNotFound indicates missing resource in required dependency path.
	ErrNotFound = errors.New("not found")
	// ErrInvalidInput indicates malformed BFF request payload.
	ErrInvalidInput = errors.New("invalid input")
)

// UpstreamError captures deterministic upstream business errors.
type UpstreamError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e UpstreamError) Error() string {
	return fmt.Sprintf("upstream status=%d code=%s", e.StatusCode, e.Code)
}

func mapDependencyError(err error) (code string, message string) {
	if err == nil {
		return "", ""
	}
	if errors.Is(err, ErrDependencyTimeout) || errors.Is(err, context.DeadlineExceeded) {
		return "dependency_timeout", "dependency call timed out"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "dependency_timeout", "dependency call timed out"
	}
	if errors.Is(err, ErrDependencyUnavailable) {
		return "dependency_unavailable", "dependency unavailable"
	}
	if errors.Is(err, ErrUnauthorized) {
		return "unauthorized", "unauthorized"
	}
	if errors.Is(err, ErrNotFound) {
		return "not_found", "resource not found"
	}
	return "upstream_error", "upstream dependency error"
}
