package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"video-streaming/backend/auth-service/internal/auth"
)

func TestValidateToken(t *testing.T) {
	t.Parallel()

	handler := NewHandler(auth.NewInMemoryValidator())
	router := handler.Router()

	tests := []struct {
		name           string
		authHeader     string
		wantStatusCode int
		wantErrorCode  string
		wantUserID     string
	}{
		{
			name:           "valid premium token",
			authHeader:     "Bearer token-premium-valid",
			wantStatusCode: http.StatusOK,
			wantUserID:     "user-premium-1",
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer unknown-token",
			wantStatusCode: http.StatusUnauthorized,
			wantErrorCode:  "unauthorized",
		},
		{
			name:           "revoked session token returns claims contract",
			authHeader:     "Bearer token-premium-revoked",
			wantStatusCode: http.StatusOK,
			wantUserID:     "user-premium-1",
		},
		{
			name:           "missing token",
			authHeader:     "",
			wantStatusCode: http.StatusUnauthorized,
			wantErrorCode:  "unauthorized",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/internal/v1/auth/validate", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatusCode {
				t.Fatalf("status mismatch: got %d want %d", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK {
				var body struct {
					UserID string `json:"userId"`
				}
				if err := json.NewDecoder(strings.NewReader(rec.Body.String())).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.UserID != tt.wantUserID {
					t.Fatalf("user mismatch: got %q want %q", body.UserID, tt.wantUserID)
				}
				return
			}

			var body struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.NewDecoder(strings.NewReader(rec.Body.String())).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Error.Code != tt.wantErrorCode {
				t.Fatalf("error code mismatch: got %q want %q", body.Error.Code, tt.wantErrorCode)
			}
		})
	}
}
