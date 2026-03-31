package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultServerHost       = "0.0.0.0"
	defaultServerPort       = 8080
	defaultReadHeaderSec    = 5
	defaultIdleTimeoutSec   = 30
	defaultShutdownTimeoutS = 10
	defaultAuthServiceURL   = "http://localhost:8081"
	defaultAuthTimeoutSec   = 2
)

// Config contains runtime settings for the API process.
type Config struct {
	ServerHost        string
	ServerPort        int
	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	AuthServiceURL    string
	AuthTimeout       time.Duration
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

	authTimeoutSec, err := intFromEnv("AUTH_TIMEOUT_SEC", defaultAuthTimeoutSec)
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_TIMEOUT_SEC: %w", err)
	}

	return Config{
		ServerHost:        stringFromEnv("SERVER_HOST", defaultServerHost),
		ServerPort:        serverPort,
		ReadHeaderTimeout: time.Duration(readHeaderSec) * time.Second,
		IdleTimeout:       time.Duration(idleTimeoutSec) * time.Second,
		ShutdownTimeout:   time.Duration(shutdownTimeoutSec) * time.Second,
		AuthServiceURL:    stringFromEnv("AUTH_SERVICE_URL", defaultAuthServiceURL),
		AuthTimeout:       time.Duration(authTimeoutSec) * time.Second,
	}, nil
}

// stringFromEnv returns the env var value or the fallback when empty.
func stringFromEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// intFromEnv parses integer env values with fallback defaults.
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
