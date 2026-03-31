package auth

import "testing"

func TestValidateClaims(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		claims  Claims
		wantErr bool
	}{
		{
			name: "valid claims",
			claims: Claims{
				UserID:       "u-1",
				Plan:         "premium",
				SessionState: SessionStateValid,
			},
			wantErr: false,
		},
		{
			name: "missing user id",
			claims: Claims{
				Plan:         "premium",
				SessionState: SessionStateValid,
			},
			wantErr: true,
		},
		{
			name: "missing plan",
			claims: Claims{
				UserID:       "u-1",
				SessionState: SessionStateValid,
			},
			wantErr: true,
		},
		{
			name: "invalid session state",
			claims: Claims{
				UserID:       "u-1",
				Plan:         "premium",
				SessionState: "revoked",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateClaims(tt.claims)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestIsPremiumPlan(t *testing.T) {
	t.Parallel()

	if !IsPremiumPlan("premium") {
		t.Fatalf("expected premium to be allowed")
	}
	if !IsPremiumPlan("PRO") {
		t.Fatalf("expected pro to be allowed")
	}
	if IsPremiumPlan("free") {
		t.Fatalf("expected free to be denied")
	}
}
