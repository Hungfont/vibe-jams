package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultServerHost        = "0.0.0.0"
	defaultServerPort        = 8085
	defaultReadHeaderSec     = 5
	defaultIdleTimeoutSec    = 30
	defaultShutdownTimeoutSec = 10
	defaultAuthServiceURL    = "http://localhost:8081"
	defaultAPIServiceURL     = "http://localhost:8084"
	defaultAuthTimeoutMS     = 800
	defaultUpstreamTimeoutMS = 5000
)

// Config contains runtime settings for api-gateway.
type Config struct {
	ServerHost        string
	ServerPort        int
	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	AuthServiceURL    string
	APIServiceURL     string
	AuthTimeout       time.Duration
	UpstreamTimeout   time.Duration
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
	idleSec, err := intFromEnv("IDLE_TIMEOUT_SEC", defaultIdleTimeoutSec)
	if err != nil {
		return Config{}, fmt.Errorf("parse IDLE_TIMEOUT_SEC: %w", err)
	}
	shutdownSec, err := intFromEnv("SHUTDOWN_TIMEOUT_SEC", defaultShutdownTimeoutSec)
	if err != nil {
		return Config{}, fmt.Errorf("parse SHUTDOWN_TIMEOUT_SEC: %w", err)
	}
	authTimeoutMS, err := intFromEnv("GATEWAY_TIMEOUT_AUTH_MS", defaultAuthTimeoutMS)
	if err != nil {
		return Config{}, fmt.Errorf("parse GATEWAY_TIMEOUT_AUTH_MS: %w", err)
	}
	upstreamTimeoutMS, err := intFromEnv("GATEWAY_TIMEOUT_UPSTREAM_MS", defaultUpstreamTimeoutMS)
	if err != nil {
		return Config{}, fmt.Errorf("parse GATEWAY_TIMEOUT_UPSTREAM_MS: %w", err)
	}

	cfg := Config{
		ServerHost:        stringFromEnv("SERVER_HOST", defaultServerHost),
		ServerPort:        serverPort,
		ReadHeaderTimeout: time.Duration(readHeaderSec) * time.Second,
		IdleTimeout:       time.Duration(idleSec) * time.Second,
		ShutdownTimeout:   time.Duration(shutdownSec) * time.Second,
		AuthServiceURL:    stringFromEnv("AUTH_SERVICE_URL", defaultAuthServiceURL),
		APIServiceURL:     stringFromEnv("API_SERVICE_URL", defaultAPIServiceURL),
		AuthTimeout:       time.Duration(authTimeoutMS) * time.Millisecond,
		UpstreamTimeout:   time.Duration(upstreamTimeoutMS) * time.Millisecond,
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate ensures required configuration values are present.
func (c Config) Validate() error {
	if c.ServerPort <= 0 {
		return fmt.Errorf("SERVER_PORT must be positive")
	}
	if c.AuthTimeout <= 0 {
		return fmt.Errorf("GATEWAY_TIMEOUT_AUTH_MS must be positive")
	}
	if c.UpstreamTimeout <= 0 {
		return fmt.Errorf("GATEWAY_TIMEOUT_UPSTREAM_MS must be positive")
	}
	if c.AuthServiceURL == "" {
		return fmt.Errorf("AUTH_SERVICE_URL must be configured")
	}
	if c.APIServiceURL == "" {
		return fmt.Errorf("API_SERVICE_URL must be configured")
	}
	return nil
}

func stringFromEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
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
