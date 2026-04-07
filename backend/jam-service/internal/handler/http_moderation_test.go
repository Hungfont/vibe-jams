package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sharedauth "video-streaming/backend/shared/auth"
)

func TestModerationEndpointsAndBlockedQueueCommand(t *testing.T) {
	t.Parallel()

	svc := newTestService()
	hostHandler := NewHTTPHandler(svc, stubValidator{claims: sharedauth.Claims{UserID: "host_1", Plan: "premium", SessionState: sharedauth.SessionStateValid}})
	memberHandler := NewHTTPHandler(svc, stubValidator{claims: sharedauth.Claims{UserID: "member_1", Plan: "free", SessionState: sharedauth.SessionStateValid}})

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/create", nil)
	createReq.Header.Set("Authorization", "Bearer token-host")
	createRec := httptest.NewRecorder()
	hostHandler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status mismatch: got %d want %d", createRec.Code, http.StatusCreated)
	}

	var created struct {
		JamID string `json:"jamId"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatalf("decode create: %v", err)
	}

	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+created.JamID+"/join", nil)
	joinReq.Header.Set("Authorization", "Bearer token-member")
	joinRec := httptest.NewRecorder()
	memberHandler.ServeHTTP(joinRec, joinReq)
	if joinRec.Code != http.StatusOK {
		t.Fatalf("join status mismatch: got %d want %d", joinRec.Code, http.StatusOK)
	}

	muteReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+created.JamID+"/moderation/mute", bytes.NewBufferString(`{"targetUserId":"member_1","reason":"spam"}`))
	muteReq.Header.Set("Authorization", "Bearer token-host")
	muteRec := httptest.NewRecorder()
	hostHandler.ServeHTTP(muteRec, muteReq)
	if muteRec.Code != http.StatusOK {
		t.Fatalf("mute status mismatch: got %d want %d", muteRec.Code, http.StatusOK)
	}

	queueReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+created.JamID+"/queue/add", bytes.NewBufferString(`{"trackId":"t1","addedBy":"member_1","idempotencyKey":"k1"}`))
	queueReq.Header.Set("Authorization", "Bearer token-member")
	queueRec := httptest.NewRecorder()
	memberHandler.ServeHTTP(queueRec, queueReq)
	if queueRec.Code != http.StatusForbidden {
		t.Fatalf("queue add status mismatch: got %d want %d", queueRec.Code, http.StatusForbidden)
	}

	var queueErr struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(queueRec.Body).Decode(&queueErr); err != nil {
		t.Fatalf("decode queue error: %v", err)
	}
	if queueErr.Error.Code != "moderation_blocked" {
		t.Fatalf("error code mismatch: got %q want moderation_blocked", queueErr.Error.Code)
	}

	nonHostMuteReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+created.JamID+"/moderation/mute", bytes.NewBufferString(`{"targetUserId":"host_1","reason":"abuse"}`))
	nonHostMuteReq.Header.Set("Authorization", "Bearer token-member")
	nonHostMuteRec := httptest.NewRecorder()
	memberHandler.ServeHTTP(nonHostMuteRec, nonHostMuteReq)
	if nonHostMuteRec.Code != http.StatusForbidden {
		t.Fatalf("non-host mute status mismatch: got %d want %d", nonHostMuteRec.Code, http.StatusForbidden)
	}

	var nonHostErr struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(nonHostMuteRec.Body).Decode(&nonHostErr); err != nil {
		t.Fatalf("decode non-host mute error: %v", err)
	}
	if nonHostErr.Error.Code != "host_only" {
		t.Fatalf("non-host mute error code mismatch: got %q want host_only", nonHostErr.Error.Code)
	}
}
