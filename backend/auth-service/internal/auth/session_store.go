package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionRevoked  = errors.New("session revoked")
	ErrSessionReused   = errors.New("session reused")
)

// RefreshSession stores lifecycle state for one opaque refresh token.
type RefreshSession struct {
	SessionID      string
	FamilyID       string
	UserID         string
	Plan           string
	SessionState   string
	Scope          []string
	TokenHash      string
	ReplacedByHash string
	ExpiresAt      time.Time
	IssuedAt       time.Time
	RevokedAt      *time.Time
	RevokeReason   string
}

// SessionStore persists refresh/session lifecycle state.
type SessionStore interface {
	Create(ctx context.Context, session RefreshSession) error
	GetByTokenHash(ctx context.Context, tokenHash string) (RefreshSession, error)
	Rotate(ctx context.Context, presentedTokenHash string, replacement RefreshSession, now time.Time) error
	RevokeFamily(ctx context.Context, familyID string, now time.Time, reason string) error
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
