package auth

import sharedauth "video-streaming/backend/shared/auth"

// DefaultFixtureClaims preserves legacy internal validation fixtures.
func DefaultFixtureClaims() map[string]sharedauth.Claims {
	return map[string]sharedauth.Claims{
		"token-premium-valid": {
			UserID:       "user-premium-1",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateValid,
		},
		"token-free-valid": {
			UserID:       "user-free-1",
			Plan:         "free",
			SessionState: sharedauth.SessionStateValid,
		},
		"token-premium-revoked": {
			UserID:       "user-premium-1",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateInvalid,
		},
	}
}
