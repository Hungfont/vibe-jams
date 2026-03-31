package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	jamauth "video-streaming/backend/jams/internal/auth"
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
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewHTTPHandler(nil, tt.validator)
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

	h := NewHTTPHandler(nil, stubValidator{err: errors.New("network down")})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jams/create", nil)
	req.Header.Set("Authorization", "Bearer token-premium-valid")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}
