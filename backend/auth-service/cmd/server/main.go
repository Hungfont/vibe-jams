package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"video-streaming/backend/auth-service/internal/auth"
	"video-streaming/backend/auth-service/internal/config"
	httpserver "video-streaming/backend/auth-service/internal/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	previousKeys, err := auth.ParsePreviousSigningKeys(cfg.JWTPreviousKeys)
	if err != nil {
		slog.Error("failed to parse previous jwt keys", "error", err)
		os.Exit(1)
	}

	keyRing, err := auth.NewKeyRing(auth.SigningKey{KeyID: cfg.JWTActiveKeyID, Secret: cfg.JWTActiveKeySecret}, previousKeys)
	if err != nil {
		slog.Error("failed to create jwt key ring", "error", err)
		os.Exit(1)
	}

	sessionStore, cleanup, err := buildSessionStore(cfg)
	if err != nil {
		slog.Error("failed to build session store", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	service, err := auth.NewService(auth.ServiceConfig{
		Credentials:     auth.NewInMemoryCredentialStore(),
		SessionStore:    sessionStore,
		KeyRing:         keyRing,
		FixtureClaims:   auth.DefaultFixtureClaims(),
		AuditLogger:     auth.NewSlogAuditLogger(slog.Default()),
		LoginLimiter:    auth.NewFixedWindowLimiter(cfg.LoginRateLimit, cfg.LoginRateLimitWindow, time.Now),
		RefreshLimiter:  auth.NewFixedWindowLimiter(cfg.RefreshRateLimit, cfg.RefreshRateLimitWindow, time.Now),
		LockoutTracker:  auth.NewLockoutTracker(cfg.LoginLockoutThreshold, cfg.LoginLockoutBackoff, time.Now),
		AccessTokenTTL:  cfg.AccessTokenTTL,
		RefreshTokenTTL: cfg.RefreshTokenTTL,
	})
	if err != nil {
		slog.Error("failed to build auth service", "error", err)
		os.Exit(1)
	}

	handler := httpserver.NewHandler(service)

	slog.Info("starting auth-service", "addr", cfg.ServerAddr, "profile", cfg.RuntimeProfile)
	if err := http.ListenAndServe(cfg.ServerAddr, handler.Router()); err != nil {
		slog.Error("auth-service stopped", "error", err)
		os.Exit(1)
	}
}

func buildSessionStore(cfg config.Config) (auth.SessionStore, func(), error) {
	switch cfg.SessionStoreBackend {
	case "inmemory":
		return auth.NewInMemorySessionStore(), func() {}, nil
	case "postgres":
		db, err := sql.Open(cfg.SessionStorePostgresDriver, cfg.SessionStorePostgresDSN)
		if err != nil {
			return nil, func() {}, fmt.Errorf("open postgres session store: %w", err)
		}
		if err := db.Ping(); err != nil {
			_ = db.Close()
			return nil, func() {}, fmt.Errorf("ping postgres session store: %w", err)
		}
		return auth.NewPostgresSessionStore(db), func() { _ = db.Close() }, nil
	default:
		return nil, func() {}, fmt.Errorf("unsupported AUTH_SESSION_STORE_BACKEND: %s", cfg.SessionStoreBackend)
	}
}
