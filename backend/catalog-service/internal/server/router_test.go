package server

import (
	"bufio"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"video-streaming/backend/catalog-service/internal/config"
	sharedcatalog "video-streaming/backend/shared/catalog"
)

func TestTrackLookupPlayable(t *testing.T) {
	t.Parallel()

	h, err := NewRouter(config.Config{RuntimeProfile: "test", CatalogBackend: "inmemory"})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
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

	h, err := NewRouter(config.Config{RuntimeProfile: "test", CatalogBackend: "inmemory"})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
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

func TestTrackLookupRestricted(t *testing.T) {
	t.Parallel()

	h, err := NewRouter(config.Config{RuntimeProfile: "test", CatalogBackend: "inmemory"})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/internal/v1/catalog/tracks/trk_3", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}

	var resp sharedcatalog.LookupResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.TrackID != "trk_3" {
		t.Fatalf("trackId mismatch: got %q", resp.TrackID)
	}
	if resp.PolicyStatus != "restricted" {
		t.Fatalf("policyStatus mismatch: got %q", resp.PolicyStatus)
	}
	if resp.PolicyReason == "" {
		t.Fatal("expected policyReason for restricted track")
	}
}

func TestTrackLookupNotFound(t *testing.T) {
	t.Parallel()

	h, err := NewRouter(config.Config{RuntimeProfile: "test", CatalogBackend: "inmemory"})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
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

func TestRequestLoggingMiddleware_PreservesFlusherAndHijacker(t *testing.T) {
	t.Parallel()

	wrapped := requestLoggingMiddleware("catalog-service", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected wrapped writer to implement http.Flusher")
		}
		flusher.Flush()

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("expected wrapped writer to implement http.Hijacker")
		}
		if _, _, err := hijacker.Hijack(); err != nil {
			t.Fatalf("expected hijack to delegate to underlying writer: %v", err)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	rec := newHijackableRecorder(t)
	wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusNoContent)
	}
	if rec.flushCount != 1 {
		t.Fatalf("flush count mismatch: got %d want 1", rec.flushCount)
	}
}

type hijackableRecorder struct {
	*httptest.ResponseRecorder
	conn       net.Conn
	rw         *bufio.ReadWriter
	flushCount int
}

func newHijackableRecorder(t *testing.T) *hijackableRecorder {
	t.Helper()

	client, server := net.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})

	return &hijackableRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		conn:             client,
		rw:               bufio.NewReadWriter(bufio.NewReader(server), bufio.NewWriter(server)),
	}
}

func (r *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.conn, r.rw, nil
}

func (r *hijackableRecorder) Flush() {
	r.flushCount++
}
