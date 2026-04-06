package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	sharedauth "video-streaming/backend/shared/auth"
)

const (
	defaultAccessTokenTTL  = 10 * time.Minute
	defaultRefreshTokenTTL = 7 * 24 * time.Hour
	opaqueTokenBytes       = 32
)

// Validator validates bearer tokens into normalized claims.
type Validator interface {
	ValidateBearerToken(token string) (sharedauth.Claims, error)
}

// Authenticator extends validation with public auth lifecycle operations.
type Authenticator interface {
	Validator
	Login(ctx context.Context, req LoginRequest) (TokenPair, error)
	Refresh(ctx context.Context, req RefreshRequest) (TokenPair, error)
	Logout(ctx context.Context, req LogoutRequest) error
	Me(ctx context.Context, accessToken string) (sharedauth.Claims, error)
}

// ServiceConfig configures service dependencies and runtime behavior.
type ServiceConfig struct {
	Credentials     CredentialStore
	SessionStore    SessionStore
	KeyRing         *KeyRing
	FixtureClaims   map[string]sharedauth.Claims
	AuditLogger     AuditLogger
	LoginLimiter    *FixedWindowLimiter
	RefreshLimiter  *FixedWindowLimiter
	LockoutTracker  *LockoutTracker
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Now             func() time.Time
	RandomReader    io.Reader
}

// Service implements public auth lifecycle and internal validation.
type Service struct {
	credentials     CredentialStore
	sessionStore    SessionStore
	keyRing         *KeyRing
	fixtureClaims   map[string]sharedauth.Claims
	audit           AuditLogger
	loginLimiter    *FixedWindowLimiter
	refreshLimiter  *FixedWindowLimiter
	lockout         *LockoutTracker
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	now             func() time.Time
	randomReader    io.Reader
}

func NewService(cfg ServiceConfig) (*Service, error) {
	if cfg.Credentials == nil {
		return nil, fmt.Errorf("credentials store is required")
	}
	if cfg.SessionStore == nil {
		return nil, fmt.Errorf("session store is required")
	}
	if cfg.KeyRing == nil {
		return nil, fmt.Errorf("key ring is required")
	}
	if cfg.AuditLogger == nil {
		cfg.AuditLogger = NoopAuditLogger{}
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.RandomReader == nil {
		cfg.RandomReader = rand.Reader
	}
	if cfg.AccessTokenTTL <= 0 {
		cfg.AccessTokenTTL = defaultAccessTokenTTL
	}
	if cfg.RefreshTokenTTL <= 0 {
		cfg.RefreshTokenTTL = defaultRefreshTokenTTL
	}
	if cfg.LoginLimiter == nil {
		cfg.LoginLimiter = NewFixedWindowLimiter(20, time.Minute, cfg.Now)
	}
	if cfg.RefreshLimiter == nil {
		cfg.RefreshLimiter = NewFixedWindowLimiter(30, time.Minute, cfg.Now)
	}
	if cfg.LockoutTracker == nil {
		cfg.LockoutTracker = NewLockoutTracker(5, 2*time.Minute, cfg.Now)
	}

	fixtures := make(map[string]sharedauth.Claims, len(cfg.FixtureClaims))
	for token, claims := range cfg.FixtureClaims {
		fixtures[token] = sharedauth.Claims{
			UserID:       claims.UserID,
			Plan:         claims.Plan,
			SessionState: claims.SessionState,
			Scope:        cloneScope(claims.Scope),
		}
	}

	return &Service{
		credentials:     cfg.Credentials,
		sessionStore:    cfg.SessionStore,
		keyRing:         cfg.KeyRing,
		fixtureClaims:   fixtures,
		audit:           cfg.AuditLogger,
		loginLimiter:    cfg.LoginLimiter,
		refreshLimiter:  cfg.RefreshLimiter,
		lockout:         cfg.LockoutTracker,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		now:             cfg.Now,
		randomReader:    cfg.RandomReader,
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (TokenPair, error) {
	identity := normalizeIdentity(req.Identity)
	if identity == "" || strings.TrimSpace(req.Password) == "" {
		s.audit.Log(ctx, AuditEvent{Event: AuditEventLoginFailure, Identity: identity, IP: req.IP, Outcome: "failure", Reason: "invalid_request", OccurredAt: s.now()})
		return TokenPair{}, ErrInvalidCredentials
	}

	if decision := s.loginLimiter.Allow(joinRateKey(req.IP, identity)); !decision.Allowed {
		s.audit.Log(ctx, AuditEvent{Event: AuditEventLoginFailure, Identity: identity, IP: req.IP, Outcome: "failure", Reason: "rate_limited", OccurredAt: s.now()})
		return TokenPair{}, ErrRateLimited
	}
	if locked, _ := s.lockout.IsLocked(identity); locked {
		s.audit.Log(ctx, AuditEvent{Event: AuditEventLoginFailure, Identity: identity, IP: req.IP, Outcome: "failure", Reason: "locked_out", OccurredAt: s.now()})
		return TokenPair{}, ErrLockedOut
	}

	user, err := s.credentials.Authenticate(ctx, identity, req.Password)
	if err != nil {
		if locked, _ := s.lockout.RegisterFailure(identity); locked {
			s.audit.Log(ctx, AuditEvent{Event: AuditEventLoginFailure, Identity: identity, IP: req.IP, Outcome: "failure", Reason: "locked_out", OccurredAt: s.now()})
			return TokenPair{}, ErrLockedOut
		}
		s.audit.Log(ctx, AuditEvent{Event: AuditEventLoginFailure, Identity: identity, IP: req.IP, Outcome: "failure", Reason: "invalid_credentials", OccurredAt: s.now()})
		return TokenPair{}, ErrInvalidCredentials
	}
	s.lockout.Reset(identity)

	now := s.now()
	sessionID, err := s.newOpaqueID()
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate session id: %w", err)
	}
	refreshToken, err := s.newOpaqueToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}
	refresh := RefreshSession{
		SessionID:    sessionID,
		FamilyID:     sessionID,
		UserID:       user.UserID,
		Plan:         user.Plan,
		SessionState: sharedauth.SessionStateValid,
		Scope:        cloneScope(user.Scope),
		TokenHash:    HashRefreshToken(refreshToken),
		ExpiresAt:    now.Add(s.refreshTokenTTL),
		IssuedAt:     now,
	}
	if err := s.sessionStore.Create(ctx, refresh); err != nil {
		return TokenPair{}, fmt.Errorf("%w: %v", ErrSessionStore, err)
	}

	pair, err := s.issueTokenPair(user.UserID, user.Plan, refresh.SessionState, user.Scope, refresh.SessionID, refreshToken, now)
	if err != nil {
		return TokenPair{}, err
	}
	s.audit.Log(ctx, AuditEvent{Event: AuditEventLoginSuccess, UserID: user.UserID, Identity: identity, IP: req.IP, SessionID: refresh.SessionID, FamilyID: refresh.FamilyID, Outcome: "success", OccurredAt: now})
	return pair, nil
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (TokenPair, error) {
	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		s.audit.Log(ctx, AuditEvent{Event: AuditEventRefreshFailure, IP: req.IP, Outcome: "failure", Reason: "missing_refresh_token", OccurredAt: s.now()})
		return TokenPair{}, ErrUnauthorized
	}

	hash := HashRefreshToken(refreshToken)
	if decision := s.refreshLimiter.Allow(joinRateKey(req.IP, hash)); !decision.Allowed {
		s.audit.Log(ctx, AuditEvent{Event: AuditEventRefreshFailure, IP: req.IP, Outcome: "failure", Reason: "rate_limited", OccurredAt: s.now()})
		return TokenPair{}, ErrRateLimited
	}

	session, err := s.sessionStore.GetByTokenHash(ctx, hash)
	if err != nil {
		s.audit.Log(ctx, AuditEvent{Event: AuditEventRefreshFailure, IP: req.IP, Outcome: "failure", Reason: "unknown_refresh_token", OccurredAt: s.now()})
		return TokenPair{}, ErrUnauthorized
	}

	now := s.now()
	if now.After(session.ExpiresAt) {
		s.audit.Log(ctx, AuditEvent{Event: AuditEventRefreshFailure, UserID: session.UserID, IP: req.IP, SessionID: session.SessionID, FamilyID: session.FamilyID, Outcome: "failure", Reason: "expired", OccurredAt: now})
		return TokenPair{}, ErrUnauthorized
	}

	nextToken, err := s.newOpaqueToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}
	replacement := RefreshSession{
		SessionID:    session.SessionID,
		FamilyID:     session.FamilyID,
		UserID:       session.UserID,
		Plan:         session.Plan,
		SessionState: session.SessionState,
		Scope:        cloneScope(session.Scope),
		TokenHash:    HashRefreshToken(nextToken),
		ExpiresAt:    now.Add(s.refreshTokenTTL),
		IssuedAt:     now,
	}
	if err := s.sessionStore.Rotate(ctx, hash, replacement, now); err != nil {
		if err == ErrSessionReused {
			_ = s.sessionStore.RevokeFamily(ctx, session.FamilyID, now, "refresh_reuse_detected")
			s.audit.Log(ctx, AuditEvent{Event: AuditEventRefreshReuse, UserID: session.UserID, IP: req.IP, SessionID: session.SessionID, FamilyID: session.FamilyID, Outcome: "failure", Reason: "reused_rotated_token", OccurredAt: now})
			return TokenPair{}, ErrRefreshReuseDetected
		}
		s.audit.Log(ctx, AuditEvent{Event: AuditEventRefreshFailure, UserID: session.UserID, IP: req.IP, SessionID: session.SessionID, FamilyID: session.FamilyID, Outcome: "failure", Reason: "invalid_or_revoked", OccurredAt: now})
		return TokenPair{}, ErrUnauthorized
	}

	pair, err := s.issueTokenPair(session.UserID, session.Plan, session.SessionState, session.Scope, session.SessionID, nextToken, now)
	if err != nil {
		return TokenPair{}, err
	}
	s.audit.Log(ctx, AuditEvent{Event: AuditEventRefreshSuccess, UserID: session.UserID, IP: req.IP, SessionID: session.SessionID, FamilyID: session.FamilyID, Outcome: "success", OccurredAt: now})
	return pair, nil
}

func (s *Service) Logout(ctx context.Context, req LogoutRequest) error {
	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		s.audit.Log(ctx, AuditEvent{Event: AuditEventLogout, IP: req.IP, Outcome: "failure", Reason: "missing_refresh_token", OccurredAt: s.now()})
		return ErrUnauthorized
	}

	hash := HashRefreshToken(refreshToken)
	session, err := s.sessionStore.GetByTokenHash(ctx, hash)
	if err != nil {
		s.audit.Log(ctx, AuditEvent{Event: AuditEventLogout, IP: req.IP, Outcome: "failure", Reason: "unknown_session", OccurredAt: s.now()})
		return ErrUnauthorized
	}

	now := s.now()
	if err := s.sessionStore.RevokeFamily(ctx, session.FamilyID, now, "logout"); err != nil {
		return fmt.Errorf("%w: %v", ErrSessionStore, err)
	}
	s.audit.Log(ctx, AuditEvent{Event: AuditEventLogout, UserID: session.UserID, IP: req.IP, SessionID: session.SessionID, FamilyID: session.FamilyID, Outcome: "success", OccurredAt: now})
	return nil
}

func (s *Service) Me(_ context.Context, accessToken string) (sharedauth.Claims, error) {
	claims, err := s.validateAccessToken(strings.TrimSpace(accessToken))
	if err != nil {
		return sharedauth.Claims{}, ErrUnauthorized
	}
	return claims, nil
}

func (s *Service) ValidateBearerToken(token string) (sharedauth.Claims, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return sharedauth.Claims{}, ErrTokenNotFound
	}
	if fixture, ok := s.fixtureClaims[trimmed]; ok {
		return fixture, nil
	}

	claims, err := s.validateAccessToken(trimmed)
	if err != nil {
		return sharedauth.Claims{}, ErrTokenNotFound
	}
	return claims, nil
}

func (s *Service) validateAccessToken(accessToken string) (sharedauth.Claims, error) {
	parsed, err := s.keyRing.VerifyAccessToken(accessToken, s.now())
	if err != nil {
		return sharedauth.Claims{}, err
	}
	normalized := parsed.NormalizedClaims()
	if err := sharedauth.ValidateClaims(normalized); err != nil {
		return sharedauth.Claims{}, err
	}
	if parsed.SessionID == "" {
		return sharedauth.Claims{}, ErrUnauthorized
	}
	return normalized, nil
}

func (s *Service) issueTokenPair(userID string, plan string, sessionState string, scope []string, sessionID string, refreshToken string, now time.Time) (TokenPair, error) {
	expiresAt := now.Add(s.accessTokenTTL)
	accessToken, err := s.keyRing.SignAccessToken(AccessTokenClaims{
		UserID:       userID,
		Plan:         plan,
		SessionState: sessionState,
		Scope:        cloneScope(scope),
		SessionID:    sessionID,
		ExpiresAt:    expiresAt.Unix(),
		IssuedAt:     now.Unix(),
	})
	if err != nil {
		return TokenPair{}, fmt.Errorf("sign access token: %w", err)
	}
	claims := sharedauth.Claims{
		UserID:       userID,
		Plan:         plan,
		SessionState: sessionState,
		Scope:        cloneScope(scope),
	}
	return TokenPair{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessExpires: expiresAt,
		Claims:        claims,
		SessionID:     sessionID,
	}, nil
}

func (s *Service) newOpaqueID() (string, error) {
	return s.newOpaqueToken()
}

func (s *Service) newOpaqueToken() (string, error) {
	bytes := make([]byte, opaqueTokenBytes)
	if _, err := io.ReadFull(s.randomReader, bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func joinRateKey(ip string, identity string) string {
	return strings.TrimSpace(ip) + "|" + strings.TrimSpace(identity)
}
