package auth

import (
	"errors"
	"strings"

	sharedauth "video-streaming/backend/shared/auth"
)

var (
	// ErrTokenNotFound indicates token is unknown/invalid.
	ErrTokenNotFound = errors.New("token not found")
)

// Validator validates tokens and produces normalized auth claims.
type Validator interface {
	ValidateBearerToken(token string) (sharedauth.Claims, error)
}

// InMemoryValidator provides deterministic auth responses for local development.
type InMemoryValidator struct {
	tokenClaims map[string]sharedauth.Claims
}

// NewInMemoryValidator builds a local validator with fixed token fixtures.
func NewInMemoryValidator() *InMemoryValidator {
	return &InMemoryValidator{
		tokenClaims: map[string]sharedauth.Claims{
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
		},
	}
}

// ValidateBearerToken returns claims when token is known; otherwise unauthorized.
func (v *InMemoryValidator) ValidateBearerToken(token string) (sharedauth.Claims, error) {
	claims, ok := v.tokenClaims[strings.TrimSpace(token)]
	if !ok {
		return sharedauth.Claims{}, ErrTokenNotFound
	}
	return claims, nil
}
