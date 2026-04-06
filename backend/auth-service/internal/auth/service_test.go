package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestServiceRefreshReuseRevokesFamily(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)
	ctx := context.Background()

	loginPair, err := svc.Login(ctx, LoginRequest{Identity: "premium@example.com", Password: "premium-pass", IP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	rotatedPair, err := svc.Refresh(ctx, RefreshRequest{RefreshToken: loginPair.RefreshToken, IP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	_, err = svc.Refresh(ctx, RefreshRequest{RefreshToken: loginPair.RefreshToken, IP: "127.0.0.1"})
	if !errors.Is(err, ErrRefreshReuseDetected) {
		t.Fatalf("expected refresh reuse error, got %v", err)
	}

	_, err = svc.Refresh(ctx, RefreshRequest{RefreshToken: rotatedPair.RefreshToken, IP: "127.0.0.1"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected family revocation unauthorized, got %v", err)
	}
}

func TestServiceLogoutRevokesSession(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)
	ctx := context.Background()

	pair, err := svc.Login(ctx, LoginRequest{Identity: "free@example.com", Password: "free-pass", IP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if err := svc.Logout(ctx, LogoutRequest{RefreshToken: pair.RefreshToken, IP: "127.0.0.1"}); err != nil {
		t.Fatalf("logout: %v", err)
	}

	_, err = svc.Refresh(ctx, RefreshRequest{RefreshToken: pair.RefreshToken, IP: "127.0.0.1"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized after logout, got %v", err)
	}
}

func TestServiceLoginFailureRateLimitLockout(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC)
	keyRing, err := NewKeyRing(SigningKey{KeyID: "kid-active", Secret: "test-secret"}, nil)
	if err != nil {
		t.Fatalf("key ring: %v", err)
	}

	svc, err := NewService(ServiceConfig{
		Credentials:     NewInMemoryCredentialStore(),
		SessionStore:    NewInMemorySessionStore(),
		KeyRing:         keyRing,
		FixtureClaims:   DefaultFixtureClaims(),
		AuditLogger:     NoopAuditLogger{},
		LoginLimiter:    NewFixedWindowLimiter(2, time.Minute, func() time.Time { return now }),
		RefreshLimiter:  NewFixedWindowLimiter(10, time.Minute, func() time.Time { return now }),
		LockoutTracker:  NewLockoutTracker(2, time.Minute, func() time.Time { return now }),
		AccessTokenTTL:  10 * time.Minute,
		RefreshTokenTTL: 30 * time.Minute,
		Now:             func() time.Time { return now },
		RandomReader:    &testRandReader{},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = svc.Login(context.Background(), LoginRequest{Identity: "premium@example.com", Password: "wrong", IP: "127.0.0.1"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}

	_, err = svc.Login(context.Background(), LoginRequest{Identity: "premium@example.com", Password: "wrong", IP: "127.0.0.1"})
	if !errors.Is(err, ErrLockedOut) {
		t.Fatalf("expected lockout, got %v", err)
	}

	_, err = svc.Login(context.Background(), LoginRequest{Identity: "premium@example.com", Password: "premium-pass", IP: "127.0.0.1"})
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected rate limited, got %v", err)
	}
}

func TestKeyRingSupportsPreviousKeys(t *testing.T) {
	t.Parallel()

	oldRing, err := NewKeyRing(SigningKey{KeyID: "kid-old", Secret: "old-secret"}, nil)
	if err != nil {
		t.Fatalf("old key ring: %v", err)
	}
	issuedAt := time.Now().UTC()
	oldToken, err := oldRing.SignAccessToken(AccessTokenClaims{
		UserID:       "u1",
		Plan:         "premium",
		SessionState: "valid",
		Scope:        []string{"jam:read"},
		SessionID:    "sid-1",
		ExpiresAt:    issuedAt.Add(5 * time.Minute).Unix(),
		IssuedAt:     issuedAt.Unix(),
	})
	if err != nil {
		t.Fatalf("sign old token: %v", err)
	}

	ring, err := NewKeyRing(SigningKey{KeyID: "kid-new", Secret: "new-secret"}, []SigningKey{{KeyID: "kid-old", Secret: "old-secret"}})
	if err != nil {
		t.Fatalf("new key ring: %v", err)
	}

	claims, err := ring.VerifyAccessToken(oldToken, issuedAt)
	if err != nil {
		t.Fatalf("verify old token with previous key: %v", err)
	}
	if claims.KeyID != "kid-old" {
		t.Fatalf("expected kid-old, got %s", claims.KeyID)
	}
}

func TestValidateBearerTokenSupportsFixturesAndJWT(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)
	fixtureClaims, err := svc.ValidateBearerToken("token-premium-valid")
	if err != nil {
		t.Fatalf("fixture validation: %v", err)
	}
	if fixtureClaims.UserID != "user-premium-1" {
		t.Fatalf("fixture userId mismatch: %s", fixtureClaims.UserID)
	}

	pair, err := svc.Login(context.Background(), LoginRequest{Identity: "premium@example.com", Password: "premium-pass", IP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	jwtClaims, err := svc.ValidateBearerToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("jwt validation: %v", err)
	}
	if len(jwtClaims.Scope) == 0 {
		t.Fatal("expected scope on jwt claims")
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()

	now := time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC)
	keyRing, err := NewKeyRing(SigningKey{KeyID: "kid-active", Secret: "test-secret"}, nil)
	if err != nil {
		t.Fatalf("new key ring: %v", err)
	}
	service, err := NewService(ServiceConfig{
		Credentials:     NewInMemoryCredentialStore(),
		SessionStore:    NewInMemorySessionStore(),
		KeyRing:         keyRing,
		FixtureClaims:   DefaultFixtureClaims(),
		AuditLogger:     NoopAuditLogger{},
		LoginLimiter:    NewFixedWindowLimiter(100, time.Minute, func() time.Time { return now }),
		RefreshLimiter:  NewFixedWindowLimiter(100, time.Minute, func() time.Time { return now }),
		LockoutTracker:  NewLockoutTracker(5, time.Minute, func() time.Time { return now }),
		AccessTokenTTL:  10 * time.Minute,
		RefreshTokenTTL: 30 * time.Minute,
		Now:             func() time.Time { return now },
		RandomReader:    &testRandReader{},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return service
}

type testRandReader struct {
	next byte
}

func (r *testRandReader) Read(p []byte) (int, error) {
	for i := range p {
		r.next++
		p[i] = r.next
	}
	return len(p), nil
}
