package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultServerHost         = "0.0.0.0"
	defaultServerPort         = 8084
	defaultReadHeaderSec      = 5
	defaultIdleTimeoutSec     = 30
	defaultShutdownTimeoutSec = 10
	defaultFeatureBFFEnabled  = true
	defaultJamServiceURL      = "http://localhost:8080"
	defaultPlaybackServiceURL = "http://localhost:8082"
	defaultCatalogServiceURL  = "http://localhost:8083"
	defaultRTGatewayURL       = "http://localhost:8086"
	defaultGatewayPublicURL   = "http://localhost:8085"
	defaultJamTimeoutMS       = 1200
	defaultPlaybackTimeoutMS  = 1000
	defaultCatalogTimeoutMS   = 800
)

// Config contains runtime settings for api-service.
type Config struct {
	ServerHost         string
	ServerPort         int
	ReadHeaderTimeout  time.Duration
	IdleTimeout        time.Duration
	ShutdownTimeout    time.Duration
	FeatureBFFEnabled  bool
	JamServiceURL      string
	PlaybackServiceURL string
	CatalogServiceURL  string
	RTGatewayURL       string
	GatewayPublicURL   string
	JamTimeout         time.Duration
	PlaybackTimeout    time.Duration
	CatalogTimeout     time.Duration
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
	jamTimeoutMS, err := intFromEnv("BFF_TIMEOUT_JAM_MS", defaultJamTimeoutMS)
	if err != nil {
		return Config{}, fmt.Errorf("parse BFF_TIMEOUT_JAM_MS: %w", err)
	}
	playbackTimeoutMS, err := intFromEnv("BFF_TIMEOUT_PLAYBACK_MS", defaultPlaybackTimeoutMS)
	if err != nil {
		return Config{}, fmt.Errorf("parse BFF_TIMEOUT_PLAYBACK_MS: %w", err)
	}
	catalogTimeoutMS, err := intFromEnv("BFF_TIMEOUT_CATALOG_MS", defaultCatalogTimeoutMS)
	if err != nil {
		return Config{}, fmt.Errorf("parse BFF_TIMEOUT_CATALOG_MS: %w", err)
	}
	featureEnabled, err := boolFromEnv("FEATURE_BFF_ENABLED", defaultFeatureBFFEnabled)
	if err != nil {
		return Config{}, fmt.Errorf("parse FEATURE_BFF_ENABLED: %w", err)
	}

	cfg := Config{
		ServerHost:         stringFromEnv("SERVER_HOST", defaultServerHost),
		ServerPort:         serverPort,
		ReadHeaderTimeout:  time.Duration(readHeaderSec) * time.Second,
		IdleTimeout:        time.Duration(idleSec) * time.Second,
		ShutdownTimeout:    time.Duration(shutdownSec) * time.Second,
		FeatureBFFEnabled:  featureEnabled,
		JamServiceURL:      stringFromEnv("JAM_SERVICE_URL", defaultJamServiceURL),
		PlaybackServiceURL: stringFromEnv("PLAYBACK_SERVICE_URL", defaultPlaybackServiceURL),
		CatalogServiceURL:  stringFromEnv("CATALOG_SERVICE_URL", defaultCatalogServiceURL),
		RTGatewayURL:       stringFromEnv("RT_GATEWAY_URL", defaultRTGatewayURL),
		GatewayPublicURL:   stringFromEnv("GATEWAY_PUBLIC_URL", defaultGatewayPublicURL),
		JamTimeout:         time.Duration(jamTimeoutMS) * time.Millisecond,
		PlaybackTimeout:    time.Duration(playbackTimeoutMS) * time.Millisecond,
		CatalogTimeout:     time.Duration(catalogTimeoutMS) * time.Millisecond,
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate ensures required URLs and timeouts are valid.
func (c Config) Validate() error {
	if c.ServerPort <= 0 {
		return fmt.Errorf("SERVER_PORT must be positive")
	}
	if c.JamTimeout <= 0 || c.PlaybackTimeout <= 0 || c.CatalogTimeout <= 0 {
		return fmt.Errorf("all BFF dependency timeouts must be positive")
	}
	if c.JamServiceURL == "" || c.PlaybackServiceURL == "" || c.CatalogServiceURL == "" || c.RTGatewayURL == "" || c.GatewayPublicURL == "" {
		return fmt.Errorf("all dependency URLs must be configured")
	}
	return nil
}

func stringFromEnv(key string, fallback string) string {
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

func boolFromEnv(key string, fallback bool) (bool, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("parse bool: %w", err)
	}
	return value, nil
}
