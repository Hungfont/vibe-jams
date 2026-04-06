package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultServerAddr            = ":8081"
	defaultRuntimeProfile        = "local"
	defaultSessionStoreBackend   = "inmemory"
	defaultSessionStoreDriver    = "postgres"
	defaultAccessTokenTTL        = 10 * time.Minute
	defaultRefreshTokenTTL       = 7 * 24 * time.Hour
	defaultLoginRateLimit        = 20
	defaultRefreshRateLimit      = 30
	defaultRateLimitWindow       = time.Minute
	defaultLoginLockoutThreshold = 5
	defaultLoginLockoutBackoff   = 2 * time.Minute
	defaultJWTActiveKeyID        = "auth-active"
)

// Config contains runtime settings for auth-service.
type Config struct {
	ServerAddr                 string
	RuntimeProfile             string
	SessionStoreBackend        string
	SessionStorePostgresDSN    string
	SessionStorePostgresDriver string
	JWTActiveKeyID             string
	JWTActiveKeySecret         string
	JWTPreviousKeys            string
	AccessTokenTTL             time.Duration
	RefreshTokenTTL            time.Duration
	LoginRateLimit             int
	LoginRateLimitWindow       time.Duration
	RefreshRateLimit           int
	RefreshRateLimitWindow     time.Duration
	LoginLockoutThreshold      int
	LoginLockoutBackoff        time.Duration
}

// Load reads runtime configuration from environment.
func Load() (Config, error) {
	loginLimit, err := intFromEnv("AUTH_LOGIN_RATE_LIMIT", defaultLoginRateLimit)
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_LOGIN_RATE_LIMIT: %w", err)
	}
	refreshLimit, err := intFromEnv("AUTH_REFRESH_RATE_LIMIT", defaultRefreshRateLimit)
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_REFRESH_RATE_LIMIT: %w", err)
	}
	lockoutThreshold, err := intFromEnv("AUTH_LOGIN_LOCKOUT_THRESHOLD", defaultLoginLockoutThreshold)
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_LOGIN_LOCKOUT_THRESHOLD: %w", err)
	}
	accessTTL, err := durationFromEnv("AUTH_ACCESS_TOKEN_TTL", defaultAccessTokenTTL)
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_ACCESS_TOKEN_TTL: %w", err)
	}
	refreshTTL, err := durationFromEnv("AUTH_REFRESH_TOKEN_TTL", defaultRefreshTokenTTL)
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_REFRESH_TOKEN_TTL: %w", err)
	}
	loginWindow, err := durationFromEnv("AUTH_LOGIN_RATE_LIMIT_WINDOW", defaultRateLimitWindow)
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_LOGIN_RATE_LIMIT_WINDOW: %w", err)
	}
	refreshWindow, err := durationFromEnv("AUTH_REFRESH_RATE_LIMIT_WINDOW", defaultRateLimitWindow)
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_REFRESH_RATE_LIMIT_WINDOW: %w", err)
	}
	lockoutBackoff, err := durationFromEnv("AUTH_LOGIN_LOCKOUT_BACKOFF", defaultLoginLockoutBackoff)
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_LOGIN_LOCKOUT_BACKOFF: %w", err)
	}

	cfg := Config{
		ServerAddr:                 stringFromEnv("SERVER_ADDR", defaultServerAddr),
		RuntimeProfile:             strings.ToLower(stringFromEnv("APP_ENV", defaultRuntimeProfile)),
		SessionStoreBackend:        strings.ToLower(stringFromEnv("AUTH_SESSION_STORE_BACKEND", defaultSessionStoreBackend)),
		SessionStorePostgresDSN:    strings.TrimSpace(os.Getenv("AUTH_SESSION_STORE_POSTGRES_DSN")),
		SessionStorePostgresDriver: strings.ToLower(stringFromEnv("AUTH_SESSION_STORE_POSTGRES_DRIVER", defaultSessionStoreDriver)),
		JWTActiveKeyID:             stringFromEnv("AUTH_JWT_ACTIVE_KID", defaultJWTActiveKeyID),
		JWTActiveKeySecret:         strings.TrimSpace(os.Getenv("AUTH_JWT_ACTIVE_SECRET")),
		JWTPreviousKeys:            strings.TrimSpace(os.Getenv("AUTH_JWT_PREVIOUS_KEYS")),
		AccessTokenTTL:             accessTTL,
		RefreshTokenTTL:            refreshTTL,
		LoginRateLimit:             loginLimit,
		LoginRateLimitWindow:       loginWindow,
		RefreshRateLimit:           refreshLimit,
		RefreshRateLimitWindow:     refreshWindow,
		LoginLockoutThreshold:      lockoutThreshold,
		LoginLockoutBackoff:        lockoutBackoff,
	}
	if err := cfg.ValidateRuntimePolicy(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// ValidateRuntimePolicy validates runtime mode restrictions.
func (c Config) ValidateRuntimePolicy() error {
	if strings.TrimSpace(c.ServerAddr) == "" {
		return fmt.Errorf("SERVER_ADDR is required")
	}
	if c.SessionStoreBackend != "inmemory" && c.SessionStoreBackend != "postgres" {
		return fmt.Errorf("invalid AUTH_SESSION_STORE_BACKEND: %s", c.SessionStoreBackend)
	}
	if !allowsInMemoryBackend(c.RuntimeProfile) && c.SessionStoreBackend == "inmemory" {
		return fmt.Errorf("AUTH_SESSION_STORE_BACKEND=inmemory is allowed only in local/dev/test profiles")
	}
	if c.SessionStoreBackend == "postgres" {
		if strings.TrimSpace(c.SessionStorePostgresDriver) == "" {
			return fmt.Errorf("AUTH_SESSION_STORE_POSTGRES_DRIVER is required when backend=postgres")
		}
		if strings.TrimSpace(c.SessionStorePostgresDSN) == "" {
			return fmt.Errorf("AUTH_SESSION_STORE_POSTGRES_DSN is required when backend=postgres")
		}
	}
	if strings.TrimSpace(c.JWTActiveKeyID) == "" {
		return fmt.Errorf("AUTH_JWT_ACTIVE_KID is required")
	}
	if strings.TrimSpace(c.JWTActiveKeySecret) == "" {
		return fmt.Errorf("AUTH_JWT_ACTIVE_SECRET is required")
	}
	if c.AccessTokenTTL <= 0 || c.RefreshTokenTTL <= 0 {
		return fmt.Errorf("token ttl values must be positive")
	}
	if c.LoginRateLimit <= 0 || c.RefreshRateLimit <= 0 {
		return fmt.Errorf("rate limits must be positive")
	}
	if c.LoginRateLimitWindow <= 0 || c.RefreshRateLimitWindow <= 0 {
		return fmt.Errorf("rate limit windows must be positive")
	}
	if c.LoginLockoutThreshold <= 0 {
		return fmt.Errorf("AUTH_LOGIN_LOCKOUT_THRESHOLD must be positive")
	}
	if c.LoginLockoutBackoff <= 0 {
		return fmt.Errorf("AUTH_LOGIN_LOCKOUT_BACKOFF must be positive")
	}
	return nil
}

func stringFromEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func intFromEnv(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("atoi: %w", err)
	}
	return value, nil
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse duration: %w", err)
	}
	return value, nil
}

func isStrictRuntimeProfile(profile string) bool {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "prod", "production", "staging":
		return true
	default:
		return false
	}
}

func allowsInMemoryBackend(profile string) bool {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "test", "local", "dev", "development":
		return true
	default:
		return false
	}
}
