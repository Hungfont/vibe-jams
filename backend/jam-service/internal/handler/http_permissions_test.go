package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sharedauth "video-streaming/backend/shared/auth"
)

func TestPermissionEndpointsReadAndUpdate(t *testing.T) {
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

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/jams/"+created.JamID+"/permissions", nil)
	getReq.Header.Set("Authorization", "Bearer token-member")
	getRec := httptest.NewRecorder()
	memberHandler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get permissions status mismatch: got %d want %d", getRec.Code, http.StatusOK)
	}

	updateReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+created.JamID+"/permissions", bytes.NewBufferString(`{"canControlPlayback":true}`))
	updateReq.Header.Set("Authorization", "Bearer token-host")
	updateRec := httptest.NewRecorder()
	hostHandler.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update permissions status mismatch: got %d want %d", updateRec.Code, http.StatusOK)
	}

	var updated struct {
		CanControlPlayback bool `json:"canControlPlayback"`
	}
	if err := json.NewDecoder(updateRec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated permissions: %v", err)
	}
	if !updated.CanControlPlayback {
		t.Fatal("expected canControlPlayback=true after host update")
	}

	nonHostUpdateReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+created.JamID+"/permissions", bytes.NewBufferString(`{"canReorderQueue":true}`))
	nonHostUpdateReq.Header.Set("Authorization", "Bearer token-member")
	nonHostUpdateRec := httptest.NewRecorder()
	memberHandler.ServeHTTP(nonHostUpdateRec, nonHostUpdateReq)
	if nonHostUpdateRec.Code != http.StatusForbidden {
		t.Fatalf("non-host update status mismatch: got %d want %d", nonHostUpdateRec.Code, http.StatusForbidden)
	}

	var nonHostErr struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(nonHostUpdateRec.Body).Decode(&nonHostErr); err != nil {
		t.Fatalf("decode non-host error: %v", err)
	}
	if nonHostErr.Error.Code != "host_only" {
		t.Fatalf("non-host error code mismatch: got %q want host_only", nonHostErr.Error.Code)
	}
}
