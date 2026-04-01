package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultServerHost       = "0.0.0.0"
	defaultServerPort       = 8083
	defaultReadHeaderSec    = 5
	defaultIdleTimeoutSec   = 30
	defaultShutdownTimeoutS = 10
	defaultRuntimeProfile   = "local"
	defaultCatalogBackend   = "inmemory"
)

// Config contains runtime settings for catalog-service.
type Config struct {
	RuntimeProfile    string
	CatalogBackend    string
	ServerHost        string
	ServerPort        int
	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

// Load reads runtime configuration from environment variables.
func Load() (Config, error) {
	serverPort, err := intFromEnv("SERVER_PORT", defaultServerPort)
	if err != nil {
		return Config{}, fmt.Errorf("parse SERVER_PORT: %w", err)
	}

	readHeaderSec, err := intFromEnv("READ_HEADER_TIMEOUT_SEC", defaultReadHeaderSec)
	if err != nil {
		return Config{}, fmt.Errorf("parse READ_HEADER_TIMEOUT_SEC: %w", err)
	}

	idleTimeoutSec, err := intFromEnv("IDLE_TIMEOUT_SEC", defaultIdleTimeoutSec)
	if err != nil {
		return Config{}, fmt.Errorf("parse IDLE_TIMEOUT_SEC: %w", err)
	}

	shutdownTimeoutSec, err := intFromEnv("SHUTDOWN_TIMEOUT_SEC", defaultShutdownTimeoutS)
	if err != nil {
		return Config{}, fmt.Errorf("parse SHUTDOWN_TIMEOUT_SEC: %w", err)
	}

	cfg := Config{
		RuntimeProfile:    strings.ToLower(stringFromEnv("APP_ENV", defaultRuntimeProfile)),
		CatalogBackend:    strings.ToLower(stringFromEnv("CATALOG_SOURCE_BACKEND", defaultCatalogBackend)),
		ServerHost:        stringFromEnv("SERVER_HOST", defaultServerHost),
		ServerPort:        serverPort,
		ReadHeaderTimeout: time.Duration(readHeaderSec) * time.Second,
		IdleTimeout:       time.Duration(idleTimeoutSec) * time.Second,
		ShutdownTimeout:   time.Duration(shutdownTimeoutSec) * time.Second,
	}

	if err := cfg.ValidateRuntimePolicy(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) ValidateRuntimePolicy() error {
	if c.CatalogBackend != "inmemory" && c.CatalogBackend != "postgres" {
		return fmt.Errorf("invalid CATALOG_SOURCE_BACKEND: %s", c.CatalogBackend)
	}
	if isStrictRuntimeProfile(c.RuntimeProfile) && c.CatalogBackend == "inmemory" {
		return fmt.Errorf("CATALOG_SOURCE_BACKEND=inmemory is allowed only in test profile")
	}
	return nil
}

func isStrictRuntimeProfile(profile string) bool {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "prod", "production", "staging":
		return true
	default:
		return false
	}
}

func stringFromEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func intFromEnv(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("atoi: %w", err)
	}
	return value, nil
}
