package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	sharedauth "video-streaming/backend/shared/auth"
)

func TestQueueRemove_StaleVersionRejectedWithRetryGuidance(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(newTestService(), stubValidator{
		claims: sharedauth.Claims{
			UserID:       "host_1",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateValid,
		},
	})

	jamID := createJamSessionForQueueTest(t, h)
	itemID := addQueueItemForQueueTest(t, h, jamID, "trk_1", "k_1")

	removeReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/jams/"+jamID+"/queue/remove",
		bytes.NewBufferString(fmt.Sprintf(`{"itemId":"%s","expectedQueueVersion":0}`, itemID)),
	)
	removeReq.Header.Set("Authorization", "Bearer token-host")
	removeRec := httptest.NewRecorder()
	h.ServeHTTP(removeRec, removeReq)

	if removeRec.Code != http.StatusConflict {
		t.Fatalf("status mismatch: got %d want %d", removeRec.Code, http.StatusConflict)
	}

	var conflict struct {
		Error struct {
			Code  string `json:"code"`
			Retry struct {
				CurrentQueueVersion int64 `json:"currentQueueVersion"`
			} `json:"retry"`
		} `json:"error"`
	}
	if err := json.NewDecoder(removeRec.Body).Decode(&conflict); err != nil {
		t.Fatalf("decode remove error: %v", err)
	}
	if conflict.Error.Code != "version_conflict" {
		t.Fatalf("error code mismatch: got %q want version_conflict", conflict.Error.Code)
	}
	if conflict.Error.Retry.CurrentQueueVersion != 1 {
		t.Fatalf("retry currentQueueVersion mismatch: got %d want 1", conflict.Error.Retry.CurrentQueueVersion)
	}

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
	if snapshot.QueueVersion != 1 || len(snapshot.Items) != 1 {
		t.Fatalf("snapshot mutated unexpectedly: version=%d items=%d", snapshot.QueueVersion, len(snapshot.Items))
	}
}

func TestQueueRemove_MatchingVersionSucceeds(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(newTestService(), stubValidator{
		claims: sharedauth.Claims{
			UserID:       "host_1",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateValid,
		},
	})

	jamID := createJamSessionForQueueTest(t, h)
	itemID := addQueueItemForQueueTest(t, h, jamID, "trk_1", "k_1")

	removeReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/jams/"+jamID+"/queue/remove",
		bytes.NewBufferString(fmt.Sprintf(`{"itemId":"%s","expectedQueueVersion":1}`, itemID)),
	)
	removeReq.Header.Set("Authorization", "Bearer token-host")
	removeRec := httptest.NewRecorder()
	h.ServeHTTP(removeRec, removeReq)

	if removeRec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", removeRec.Code, http.StatusOK)
	}

	var snapshot struct {
		QueueVersion int64 `json:"queueVersion"`
		Items        []any `json:"items"`
	}
	if err := json.NewDecoder(removeRec.Body).Decode(&snapshot); err != nil {
		t.Fatalf("decode remove response: %v", err)
	}
	if snapshot.QueueVersion != 2 {
		t.Fatalf("queueVersion mismatch: got %d want 2", snapshot.QueueVersion)
	}
	if len(snapshot.Items) != 0 {
		t.Fatalf("queue items mismatch: got %d want 0", len(snapshot.Items))
	}
}

func TestQueueReorder_StaleVersionRejectedWithRetryGuidance(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(newTestService(), stubValidator{
		claims: sharedauth.Claims{
			UserID:       "host_1",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateValid,
		},
	})

	jamID := createJamSessionForQueueTest(t, h)
	firstItem := addQueueItemForQueueTest(t, h, jamID, "trk_1", "k_1")
	secondItem := addQueueItemForQueueTest(t, h, jamID, "trk_2", "k_2")

	reorderReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/jams/"+jamID+"/queue/reorder",
		bytes.NewBufferString(fmt.Sprintf(`{"itemIds":["%s","%s"],"expectedQueueVersion":1}`, secondItem, firstItem)),
	)
	reorderReq.Header.Set("Authorization", "Bearer token-host")
	reorderRec := httptest.NewRecorder()
	h.ServeHTTP(reorderRec, reorderReq)

	if reorderRec.Code != http.StatusConflict {
		t.Fatalf("status mismatch: got %d want %d", reorderRec.Code, http.StatusConflict)
	}

	var conflict struct {
		Error struct {
			Code  string `json:"code"`
			Retry struct {
				CurrentQueueVersion int64 `json:"currentQueueVersion"`
			} `json:"retry"`
		} `json:"error"`
	}
	if err := json.NewDecoder(reorderRec.Body).Decode(&conflict); err != nil {
		t.Fatalf("decode reorder error: %v", err)
	}
	if conflict.Error.Code != "version_conflict" {
		t.Fatalf("error code mismatch: got %q want version_conflict", conflict.Error.Code)
	}
	if conflict.Error.Retry.CurrentQueueVersion != 2 {
		t.Fatalf("retry currentQueueVersion mismatch: got %d want 2", conflict.Error.Retry.CurrentQueueVersion)
	}
}

func createJamSessionForQueueTest(t *testing.T, h *HTTPHandler) string {
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

func addQueueItemForQueueTest(t *testing.T, h *HTTPHandler, jamID string, trackID string, idempotencyKey string) string {
	t.Helper()

	addReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/jams/"+jamID+"/queue/add",
		bytes.NewBufferString(fmt.Sprintf(`{"trackId":"%s","addedBy":"host_1","idempotencyKey":"%s"}`, trackID, idempotencyKey)),
	)
	addReq.Header.Set("Authorization", "Bearer token-host")
	addRec := httptest.NewRecorder()
	h.ServeHTTP(addRec, addReq)

	if addRec.Code != http.StatusOK {
		t.Fatalf("add status mismatch: got %d want %d", addRec.Code, http.StatusOK)
	}

	var snapshot struct {
		Items []struct {
			ItemID string `json:"itemId"`
		} `json:"items"`
	}
	if err := json.NewDecoder(addRec.Body).Decode(&snapshot); err != nil {
		t.Fatalf("decode add response: %v", err)
	}
	if len(snapshot.Items) == 0 {
		t.Fatal("expected at least one queue item")
	}

	return snapshot.Items[len(snapshot.Items)-1].ItemID
}
