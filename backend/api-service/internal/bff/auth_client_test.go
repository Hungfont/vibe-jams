package bff

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPAuthClientValidateBearerToken_ValidClaims(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"userId":"user_1","plan":"premium","sessionState":"valid"}`))
	}))
	defer ts.Close()

	client := NewHTTPAuthClient(ts.URL, 500*time.Millisecond)
	claims, err := client.ValidateBearerToken(context.Background(), "Bearer token")
	if err != nil {
		t.Fatalf("ValidateBearerToken() error = %v", err)
	}
	if claims.UserID != "user_1" || claims.Plan != "premium" || claims.SessionState != "valid" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestHTTPAuthClientValidateBearerToken_InvalidContractRejected(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"userId":"","plan":"premium","sessionState":"valid"}`))
	}))
	defer ts.Close()

	client := NewHTTPAuthClient(ts.URL, 500*time.Millisecond)
	_, err := client.ValidateBearerToken(context.Background(), "Bearer token")
	if err != ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
