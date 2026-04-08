package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"video-streaming/backend/jams/internal/repository"
	"video-streaming/backend/jams/internal/service"
	sharedauth "video-streaming/backend/shared/auth"
	sharedcatalog "video-streaming/backend/shared/catalog"
)

type stubCatalogValidator struct {
	err  error
	resp sharedcatalog.LookupResponse
}

func (s stubCatalogValidator) ValidateTrack(_ context.Context, _ string) (sharedcatalog.LookupResponse, error) {
	if s.err != nil {
		return sharedcatalog.LookupResponse{}, s.err
	}
	if s.resp.TrackID != "" {
		return s.resp, nil
	}
	return sharedcatalog.LookupResponse{TrackID: "trk_ok", IsPlayable: true}, nil
}

func TestAddQueueTrackRejectedWhenCatalogNotFound(t *testing.T) {
	t.Parallel()

	h := newCatalogValidationHandler(stubCatalogValidator{err: sharedcatalog.ErrTrackNotFound}, true)
	jamID := mustCreateSession(t, h)

	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+jamID+"/queue/add", bytes.NewBufferString(`{"trackId":"trk_missing","addedBy":"host_1","idempotencyKey":"k_nf"}`))
	addReq.Header.Set("Authorization", "Bearer token-host")
	addRec := httptest.NewRecorder()
	h.ServeHTTP(addRec, addReq)

	if addRec.Code != http.StatusNotFound {
		t.Fatalf("status mismatch: got %d want %d", addRec.Code, http.StatusNotFound)
	}
	assertErrorCode(t, addRec, "track_not_found")
	assertSnapshotUnchanged(t, h, jamID)
}

func TestAddQueueTrackRejectedWhenCatalogUnavailable(t *testing.T) {
	t.Parallel()

	h := newCatalogValidationHandler(stubCatalogValidator{err: sharedcatalog.ErrTrackUnavailable}, true)
	jamID := mustCreateSession(t, h)

	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+jamID+"/queue/add", bytes.NewBufferString(`{"trackId":"trk_blocked","addedBy":"host_1","idempotencyKey":"k_un"}`))
	addReq.Header.Set("Authorization", "Bearer token-host")
	addRec := httptest.NewRecorder()
	h.ServeHTTP(addRec, addReq)

	if addRec.Code != http.StatusConflict {
		t.Fatalf("status mismatch: got %d want %d", addRec.Code, http.StatusConflict)
	}
	assertErrorCode(t, addRec, "track_unavailable")
	assertSnapshotUnchanged(t, h, jamID)
}

func TestAddQueueTrackAcceptedWhenCatalogPlayable(t *testing.T) {
	t.Parallel()

	h := newCatalogValidationHandler(stubCatalogValidator{}, true)
	jamID := mustCreateSession(t, h)

	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+jamID+"/queue/add", bytes.NewBufferString(`{"trackId":"trk_ok","addedBy":"host_1","idempotencyKey":"k_ok"}`))
	addReq.Header.Set("Authorization", "Bearer token-host")
	addRec := httptest.NewRecorder()
	h.ServeHTTP(addRec, addReq)
	if addRec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", addRec.Code, http.StatusOK)
	}

	var snap struct {
		QueueVersion int64 `json:"queueVersion"`
		Items        []struct {
			TrackID string `json:"trackId"`
		} `json:"items"`
	}
	if err := json.NewDecoder(addRec.Body).Decode(&snap); err != nil {
		t.Fatalf("decode add response: %v", err)
	}
	if snap.QueueVersion != 1 {
		t.Fatalf("queueVersion mismatch: got %d want 1", snap.QueueVersion)
	}
	if len(snap.Items) != 1 || snap.Items[0].TrackID != "trk_ok" {
		t.Fatalf("unexpected items in snapshot: %+v", snap.Items)
	}
}

func TestAddQueueTrackRejectedWhenCatalogRestricted(t *testing.T) {
	t.Parallel()

	h := newCatalogValidationHandler(stubCatalogValidator{
		resp: sharedcatalog.LookupResponse{
			TrackID:      "trk_3",
			IsPlayable:   true,
			PolicyStatus: "restricted",
			PolicyReason: "region_blocked",
		},
	}, true)
	jamID := mustCreateSession(t, h)

	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+jamID+"/queue/add", bytes.NewBufferString(`{"trackId":"trk_3","addedBy":"host_1","idempotencyKey":"k_rs"}`))
	addReq.Header.Set("Authorization", "Bearer token-host")
	addRec := httptest.NewRecorder()
	h.ServeHTTP(addRec, addReq)

	if addRec.Code != http.StatusForbidden {
		t.Fatalf("status mismatch: got %d want %d", addRec.Code, http.StatusForbidden)
	}
	assertErrorCode(t, addRec, "track_restricted")
	assertSnapshotUnchanged(t, h, jamID)
}

func TestAddQueueTrackPolicyOffAllowsRestrictedTrack(t *testing.T) {
	t.Parallel()

	h := newCatalogValidationHandler(stubCatalogValidator{
		resp: sharedcatalog.LookupResponse{
			TrackID:      "trk_3",
			IsPlayable:   true,
			PolicyStatus: "restricted",
			PolicyReason: "region_blocked",
		},
	}, false)
	jamID := mustCreateSession(t, h)

	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+jamID+"/queue/add", bytes.NewBufferString(`{"trackId":"trk_3","addedBy":"host_1","idempotencyKey":"k_policy_off"}`))
	addReq.Header.Set("Authorization", "Bearer token-host")
	addRec := httptest.NewRecorder()
	h.ServeHTTP(addRec, addReq)
	if addRec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", addRec.Code, http.StatusOK)
	}
}

func newCatalogValidationHandler(catalogValidator stubCatalogValidator, enabled bool) *HTTPHandler {
	repo := repository.NewRedisQueueRepository()
	svc := service.NewWithCatalogValidator(repo, nil, catalogValidator, enabled)
	return NewHTTPHandler(svc, stubValidator{
		claims: sharedauth.Claims{
			UserID:       "host_1",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateValid,
		},
	})
}

func mustCreateSession(t *testing.T, h *HTTPHandler) string {
	t.Helper()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/create", nil)
	createReq.Header.Set("Authorization", "Bearer token-host")
	createRec := httptest.NewRecorder()
	h.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status mismatch: got %d want %d", createRec.Code, http.StatusCreated)
	}

	var created struct {
		JamID string `json:"jamId"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.JamID == "" {
		t.Fatal("expected jamId in create response")
	}
	return created.JamID
}

func assertSnapshotUnchanged(t *testing.T, h *HTTPHandler, jamID string) {
	t.Helper()
	snapshotReq := httptest.NewRequest(http.MethodGet, "/api/v1/jams/"+jamID+"/queue/snapshot", nil)
	snapshotRec := httptest.NewRecorder()
	h.ServeHTTP(snapshotRec, snapshotReq)
	if snapshotRec.Code != http.StatusOK {
		t.Fatalf("snapshot status mismatch: got %d want %d", snapshotRec.Code, http.StatusOK)
	}
	var snapshot struct {
		QueueVersion int64 `json:"queueVersion"`
		Items        []any `json:"items"`
	}
	if err := json.NewDecoder(snapshotRec.Body).Decode(&snapshot); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snapshot.QueueVersion != 0 {
		t.Fatalf("queueVersion mutated unexpectedly: got %d want 0", snapshot.QueueVersion)
	}
	if len(snapshot.Items) != 0 {
		t.Fatalf("queue items mutated unexpectedly: got %d want 0", len(snapshot.Items))
	}
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error.Code != want {
		t.Fatalf("error code mismatch: got %q want %q", body.Error.Code, want)
	}
}
