package auth

import (
	"errors"
	"time"

	sharedauth "video-streaming/backend/shared/auth"
)

var (
	ErrTokenNotFound        = errors.New("token not found")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrRateLimited          = errors.New("rate limited")
	ErrLockedOut            = errors.New("locked out")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrRefreshReuseDetected = errors.New("refresh reuse detected")
	ErrSessionStore         = errors.New("session store error")
)

// LoginRequest contains public login endpoint inputs.
type LoginRequest struct {
	Identity string
	Password string
	IP       string
}

// RefreshRequest contains public refresh endpoint inputs.
type RefreshRequest struct {
	RefreshToken string
	IP           string
}

// LogoutRequest contains public logout endpoint inputs.
type LogoutRequest struct {
	RefreshToken string
	IP           string
}

// TokenPair contains issued access and refresh credentials.
type TokenPair struct {
	AccessToken   string
	RefreshToken  string
	AccessExpires time.Time
	Claims        sharedauth.Claims
	SessionID     string
}

// UserRecord represents an authenticated principal.
type UserRecord struct {
	UserID string
	Plan   string
	Scope  []string
}

// AccessTokenClaims captures fields persisted in issued JWT tokens.
type AccessTokenClaims struct {
	UserID       string   `json:"userId"`
	Plan         string   `json:"plan"`
	SessionState string   `json:"sessionState"`
	Scope        []string `json:"scope,omitempty"`
	SessionID    string   `json:"sid"`
	KeyID        string   `json:"kid"`
	ExpiresAt    int64    `json:"exp"`
	IssuedAt     int64    `json:"iat"`
}

func (c AccessTokenClaims) NormalizedClaims() sharedauth.Claims {
	return sharedauth.Claims{
		UserID:       c.UserID,
		Plan:         c.Plan,
		SessionState: c.SessionState,
		Scope:        cloneScope(c.Scope),
	}
}

func cloneScope(scope []string) []string {
	if len(scope) == 0 {
		return nil
	}
	cloned := make([]string, 0, len(scope))
	for _, item := range scope {
		if item == "" {
			continue
		}
		cloned = append(cloned, item)
	}
	if len(cloned) == 0 {
		return nil
	}
	return cloned
}
