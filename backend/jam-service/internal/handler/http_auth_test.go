package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	jamauth "video-streaming/backend/jams/internal/auth"
	"video-streaming/backend/jams/internal/repository"
	"video-streaming/backend/jams/internal/service"
	sharedauth "video-streaming/backend/shared/auth"
)

type stubValidator struct {
	claims sharedauth.Claims
	err    error
}

func (s stubValidator) ValidateBearerToken(_ context.Context, _ string) (sharedauth.Claims, error) {
	if s.err != nil {
		return sharedauth.Claims{}, s.err
	}
	return s.claims, nil
}

func TestCreateAndEndAuthorizationMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		authHeader     string
		validator      stubValidator
		wantStatusCode int
		wantErrorCode  string
	}{
		{
			name:           "missing token returns unauthorized",
			path:           "/api/v1/jams/create",
			wantStatusCode: http.StatusUnauthorized,
			wantErrorCode:  "unauthorized",
		},
		{
			name:       "invalid token returns unauthorized",
			path:       "/api/v1/jams/create",
			authHeader: "Bearer bad-token",
			validator: stubValidator{
				err: jamauth.ErrUnauthorized,
			},
			wantStatusCode: http.StatusUnauthorized,
			wantErrorCode:  "unauthorized",
		},
		{
			name:       "invalid session returns unauthorized",
			path:       "/api/v1/jams/create",
			authHeader: "Bearer token-premium-revoked",
			validator: stubValidator{
				claims: sharedauth.Claims{
					UserID:       "user-1",
					Plan:         "premium",
					SessionState: sharedauth.SessionStateInvalid,
				},
			},
			wantStatusCode: http.StatusUnauthorized,
			wantErrorCode:  "unauthorized",
		},
		{
			name:       "non premium returns forbidden premium_required",
			path:       "/api/v1/jams/create",
			authHeader: "Bearer token-free-valid",
			validator: stubValidator{
				claims: sharedauth.Claims{
					UserID:       "user-free-1",
					Plan:         "free",
					SessionState: sharedauth.SessionStateValid,
				},
			},
			wantStatusCode: http.StatusForbidden,
			wantErrorCode:  "premium_required",
		},
		{
			name:       "premium can create",
			path:       "/api/v1/jams/create",
			authHeader: "Bearer token-premium-valid",
			validator: stubValidator{
				claims: sharedauth.Claims{
					UserID:       "user-premium-1",
					Plan:         "premium",
					SessionState: sharedauth.SessionStateValid,
				},
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:       "premium can end",
			path:       "/api/v1/jams/jam-1/end",
			authHeader: "Bearer token-premium-valid",
			validator: stubValidator{
				claims: sharedauth.Claims{
					UserID:       "user-premium-1",
					Plan:         "premium",
					SessionState: sharedauth.SessionStateValid,
				},
			},
			wantStatusCode: http.StatusNotFound,
			wantErrorCode:  "not_found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newTestHandler(tt.validator)
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatusCode {
				t.Fatalf("status mismatch: got %d want %d", rec.Code, tt.wantStatusCode)
			}
			if tt.wantErrorCode == "" {
				return
			}

			var body struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Error.Code != tt.wantErrorCode {
				t.Fatalf("error code mismatch: got %q want %q", body.Error.Code, tt.wantErrorCode)
			}
		})
	}
}

func TestCreateAuthorizationHandlesUnknownErrorAsUnauthorized(t *testing.T) {
	t.Parallel()

	h := newTestHandler(stubValidator{err: errors.New("network down")})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jams/create", nil)
	req.Header.Set("Authorization", "Bearer token-premium-valid")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestSessionLifecycleJoinLeaveAndHostLeaveEnds(t *testing.T) {
	t.Parallel()

	svc := newTestService()
	h := NewHTTPHandler(svc, stubValidator{
		claims: sharedauth.Claims{
			UserID:       "host_1",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateValid,
		},
	})

	// create
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
		t.Fatalf("decode create: %v", err)
	}
	if created.JamID == "" {
		t.Fatal("expected jamId in create response")
	}

	// join as member
	joinHandler := NewHTTPHandler(svc, stubValidator{
		claims: sharedauth.Claims{
			UserID:       "member_1",
			Plan:         "free",
			SessionState: sharedauth.SessionStateValid,
		},
	})
	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+created.JamID+"/join", nil)
	joinReq.Header.Set("Authorization", "Bearer token-member")
	joinRec := httptest.NewRecorder()
	joinHandler.ServeHTTP(joinRec, joinReq)
	if joinRec.Code != http.StatusOK {
		t.Fatalf("join status mismatch: got %d want %d", joinRec.Code, http.StatusOK)
	}

	// host leave ends session
	leaveReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+created.JamID+"/leave", nil)
	leaveReq.Header.Set("Authorization", "Bearer token-host")
	leaveRec := httptest.NewRecorder()
	h.ServeHTTP(leaveRec, leaveReq)
	if leaveRec.Code != http.StatusOK {
		t.Fatalf("leave status mismatch: got %d want %d", leaveRec.Code, http.StatusOK)
	}
	var leaveBody struct {
		Status   string `json:"status"`
		EndCause string `json:"endCause"`
	}
	if err := json.NewDecoder(leaveRec.Body).Decode(&leaveBody); err != nil {
		t.Fatalf("decode leave: %v", err)
	}
	if leaveBody.Status != "ended" {
		t.Fatalf("leave status mismatch: got %q want ended", leaveBody.Status)
	}
	if leaveBody.EndCause != "host_leave" {
		t.Fatalf("leave endCause mismatch: got %q want host_leave", leaveBody.EndCause)
	}

	// queue write should now be blocked
	queueReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+created.JamID+"/queue/add", bytes.NewBufferString(`{"trackId":"t1","addedBy":"host_1","idempotencyKey":"k1"}`))
	queueRec := httptest.NewRecorder()
	h.ServeHTTP(queueRec, queueReq)
	if queueRec.Code != http.StatusConflict {
		t.Fatalf("queue add status mismatch: got %d want %d", queueRec.Code, http.StatusConflict)
	}
	var queueErr struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(queueRec.Body).Decode(&queueErr); err != nil {
		t.Fatalf("decode queue error: %v", err)
	}
	if queueErr.Error.Code != "session_ended" {
		t.Fatalf("error code mismatch: got %q want session_ended", queueErr.Error.Code)
	}
}

func TestEndRequiresHostOwnership(t *testing.T) {
	t.Parallel()

	svc := newTestService()
	hostHandler := NewHTTPHandler(svc, stubValidator{
		claims: sharedauth.Claims{
			UserID:       "host_1",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateValid,
		},
	})
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

	memberHandler := NewHTTPHandler(svc, stubValidator{
		claims: sharedauth.Claims{
			UserID:       "member_1",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateValid,
		},
	})
	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+created.JamID+"/join", nil)
	joinReq.Header.Set("Authorization", "Bearer token-member")
	joinRec := httptest.NewRecorder()
	memberHandler.ServeHTTP(joinRec, joinReq)
	if joinRec.Code != http.StatusOK {
		t.Fatalf("join status mismatch: got %d want %d", joinRec.Code, http.StatusOK)
	}

	endReq := httptest.NewRequest(http.MethodPost, "/api/v1/jams/"+created.JamID+"/end", nil)
	endReq.Header.Set("Authorization", "Bearer token-member")
	endRec := httptest.NewRecorder()
	memberHandler.ServeHTTP(endRec, endReq)
	if endRec.Code != http.StatusForbidden {
		t.Fatalf("end status mismatch: got %d want %d", endRec.Code, http.StatusForbidden)
	}
}

func newTestHandler(validator stubValidator) *HTTPHandler {
	return NewHTTPHandler(newTestService(), validator)
}

func newTestService() *service.Service {
	repo := repository.NewRedisQueueRepository()
	return service.New(repo, nil)
}
