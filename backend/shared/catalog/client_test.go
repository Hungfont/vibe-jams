package catalog

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPValidatorValidateTrackPlayable(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trackId":"trk_1","isPlayable":true,"title":"Song A","artist":"Artist A"}`))
	}))
	defer ts.Close()

	validator := NewHTTPValidator(ts.URL, 2*time.Second)
	result, err := validator.ValidateTrack(context.Background(), "trk_1")
	if err != nil {
		t.Fatalf("ValidateTrack() error = %v", err)
	}
	if result.TrackID != "trk_1" {
		t.Fatalf("trackId mismatch: got %q", result.TrackID)
	}
	if !result.IsPlayable {
		t.Fatal("expected playable=true")
	}
}

func TestHTTPValidatorValidateTrackUnavailable(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trackId":"trk_2","isPlayable":false,"reasonCode":"license_blocked"}`))
	}))
	defer ts.Close()

	validator := NewHTTPValidator(ts.URL, 2*time.Second)
	result, err := validator.ValidateTrack(context.Background(), "trk_2")
	if !errors.Is(err, ErrTrackUnavailable) {
		t.Fatalf("expected ErrTrackUnavailable, got %v", err)
	}
	if result.ReasonCode != "license_blocked" {
		t.Fatalf("reasonCode mismatch: got %q", result.ReasonCode)
	}
}

func TestHTTPValidatorValidateTrackNotFound(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	validator := NewHTTPValidator(ts.URL, 2*time.Second)
	_, err := validator.ValidateTrack(context.Background(), "missing")
	if !errors.Is(err, ErrTrackNotFound) {
		t.Fatalf("expected ErrTrackNotFound, got %v", err)
	}
}
