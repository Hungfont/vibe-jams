package bff

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPCatalogClientLookupTrack_SchemaCompatibility(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trackId":"trk_1","isPlayable":true,"title":"Song","artist":"Artist"}`))
	}))
	defer ts.Close()

	client := NewHTTPCatalogClient(ts.URL, 500*time.Millisecond)
	track, err := client.LookupTrack(context.Background(), "trk_1")
	if err != nil {
		t.Fatalf("LookupTrack() error = %v", err)
	}
	if track.TrackID != "trk_1" || !track.IsPlayable {
		t.Fatalf("unexpected track response: %+v", track)
	}
}

func TestHTTPCatalogClientLookupTrack_NotFoundMapped(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"code":"track_not_found","message":"track not found"}}`))
	}))
	defer ts.Close()

	client := NewHTTPCatalogClient(ts.URL, 500*time.Millisecond)
	_, err := client.LookupTrack(context.Background(), "trk_missing")
	up, ok := err.(UpstreamError)
	if !ok {
		t.Fatalf("expected UpstreamError, got %T %v", err, err)
	}
	if up.Code != "track_not_found" {
		t.Fatalf("unexpected upstream code: %s", up.Code)
	}
}
