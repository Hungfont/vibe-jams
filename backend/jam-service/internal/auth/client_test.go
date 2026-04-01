package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPValidatorReturnsDependencyUnavailableOnServerError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	validator := NewHTTPValidator(ts.URL, 2*time.Second)
	_, err := validator.ValidateBearerToken(context.Background(), "Bearer token-x")
	if !errors.Is(err, ErrDependencyUnavailable) {
		t.Fatalf("expected ErrDependencyUnavailable, got %v", err)
	}
}
