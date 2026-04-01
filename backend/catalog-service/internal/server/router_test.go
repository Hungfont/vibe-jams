package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sharedcatalog "video-streaming/backend/shared/catalog"
)

func TestTrackLookupPlayable(t *testing.T) {
	t.Parallel()

	h := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/internal/v1/catalog/tracks/trk_1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}

	var resp sharedcatalog.LookupResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.TrackID != "trk_1" {
		t.Fatalf("trackId mismatch: got %q", resp.TrackID)
	}
	if !resp.IsPlayable {
		t.Fatal("expected playable track")
	}
}

func TestTrackLookupUnavailable(t *testing.T) {
	t.Parallel()

	h := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/internal/v1/catalog/tracks/trk_2", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}

	var resp sharedcatalog.LookupResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.TrackID != "trk_2" {
		t.Fatalf("trackId mismatch: got %q", resp.TrackID)
	}
	if resp.IsPlayable {
		t.Fatal("expected unavailable track")
	}
	if resp.ReasonCode == "" {
		t.Fatal("expected reasonCode for unavailable track")
	}
}

func TestTrackLookupNotFound(t *testing.T) {
	t.Parallel()

	h := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/internal/v1/catalog/tracks/trk_missing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusNotFound)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error.Code != "track_not_found" {
		t.Fatalf("error code mismatch: got %q want track_not_found", body.Error.Code)
	}
}
