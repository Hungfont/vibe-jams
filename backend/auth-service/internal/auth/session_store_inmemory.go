package auth

import (
	"context"
	"sync"
	"time"
)

// InMemorySessionStore is deterministic and suitable for tests/local execution.
type InMemorySessionStore struct {
	mu      sync.Mutex
	byToken map[string]RefreshSession
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{byToken: make(map[string]RefreshSession)}
}

func (s *InMemorySessionStore) Create(_ context.Context, session RefreshSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byToken[session.TokenHash] = cloneSession(session)
	return nil
}

func (s *InMemorySessionStore) GetByTokenHash(_ context.Context, tokenHash string) (RefreshSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.byToken[tokenHash]
	if !ok {
		return RefreshSession{}, ErrSessionNotFound
	}
	return cloneSession(session), nil
}

func (s *InMemorySessionStore) Rotate(_ context.Context, presentedTokenHash string, replacement RefreshSession, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.byToken[presentedTokenHash]
	if !ok {
		return ErrSessionNotFound
	}
	if session.RevokedAt != nil {
		if session.ReplacedByHash != "" {
			return ErrSessionReused
		}
		return ErrSessionRevoked
	}
	if now.After(session.ExpiresAt) {
		return ErrSessionExpired
	}

	revokedAt := now
	session.RevokedAt = &revokedAt
	session.RevokeReason = "rotated"
	session.ReplacedByHash = replacement.TokenHash
	s.byToken[presentedTokenHash] = session

	s.byToken[replacement.TokenHash] = cloneSession(replacement)
	return nil
}

func (s *InMemorySessionStore) RevokeFamily(_ context.Context, familyID string, now time.Time, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for tokenHash, session := range s.byToken {
		if session.FamilyID != familyID {
			continue
		}
		if session.RevokedAt != nil {
			continue
		}
		revokedAt := now
		session.RevokedAt = &revokedAt
		session.RevokeReason = reason
		s.byToken[tokenHash] = session
	}
	return nil
}

func cloneSession(session RefreshSession) RefreshSession {
	cloned := RefreshSession{
		SessionID:      session.SessionID,
		FamilyID:       session.FamilyID,
		UserID:         session.UserID,
		Plan:           session.Plan,
		SessionState:   session.SessionState,
		Scope:          cloneScope(session.Scope),
		TokenHash:      session.TokenHash,
		ReplacedByHash: session.ReplacedByHash,
		ExpiresAt:      session.ExpiresAt,
		IssuedAt:       session.IssuedAt,
		RevokeReason:   session.RevokeReason,
	}
	if session.RevokedAt != nil {
		revokedAt := *session.RevokedAt
		cloned.RevokedAt = &revokedAt
	}
	return cloned
}
