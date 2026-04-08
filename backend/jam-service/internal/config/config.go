package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultServerHost              = "0.0.0.0"
	defaultServerPort              = 8080
	defaultReadHeaderSec           = 5
	defaultIdleTimeoutSec          = 30
	defaultShutdownTimeoutS        = 10
	defaultAuthServiceURL          = "http://localhost:8081"
	defaultAuthTimeoutSec          = 2
	defaultCatalogServiceURL       = "http://localhost:8083"
	defaultCatalogTimeoutSec       = 1
	defaultEnableCatalogValidation = false
	defaultRuntimeProfile          = "local"
	defaultKafkaTransport          = "inmemory"
	defaultKafkaBootstrapServers   = "localhost:9092"
	defaultStateStoreBackend       = "redis"
	defaultStateStorePathLocal     = ".runtime/jam-state.json"
	defaultAuthValidationBackend   = "http"
	defaultJWTActiveKeyID          = "auth-active"
)

// Config contains runtime settings for the API process.
type Config struct {
	RuntimeProfile          string
	KafkaTransport          string
	KafkaBootstrapServers   string
	StateStoreBackend       string
	StateStorePath          string
	AuthValidationBackend   string
	ServerHost              string
	ServerPort              int
	ReadHeaderTimeout       time.Duration
	IdleTimeout             time.Duration
	ShutdownTimeout         time.Duration
	AuthServiceURL          string
	AuthTimeout             time.Duration
	CatalogServiceURL       string
	CatalogTimeout          time.Duration
	EnableCatalogValidation bool
	JWTActiveKeyID          string
	JWTActiveKeySecret      string
	JWTPreviousKeys         string
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

	catalogTimeoutSec, err := intFromEnv("CATALOG_TIMEOUT_SEC", defaultCatalogTimeoutSec)
	if err != nil {
		return Config{}, fmt.Errorf("parse CATALOG_TIMEOUT_SEC: %w", err)
	}

	enableCatalogValidation, err := boolFromEnv("ENABLE_CATALOG_VALIDATION", defaultEnableCatalogValidation)
	if err != nil {
		return Config{}, fmt.Errorf("parse ENABLE_CATALOG_VALIDATION: %w", err)
	}

	runtimeProfile := strings.ToLower(stringFromEnv("APP_ENV", defaultRuntimeProfile))
	kafkaTransport := strings.ToLower(stringFromEnv("KAFKA_TRANSPORT", defaultKafkaTransport))
	stateStoreBackend := strings.ToLower(stringFromEnv("STATE_STORE_BACKEND", defaultStateStoreBackend))
	stateStorePath := strings.TrimSpace(os.Getenv("STATE_STORE_PATH"))
	if stateStorePath == "" && runtimeProfile == "local" && stateStoreBackend != "inmemory" {
		stateStorePath = defaultStateStorePathLocal
	}
	kafkaBootstrap := strings.TrimSpace(os.Getenv("KAFKA_BOOTSTRAP_SERVERS"))
	if kafkaBootstrap == "" && runtimeProfile == "local" {
		kafkaBootstrap = defaultKafkaBootstrapServers
	}

	cfg := Config{
		RuntimeProfile:        runtimeProfile,
		KafkaTransport:        kafkaTransport,
		KafkaBootstrapServers: kafkaBootstrap,
		StateStoreBackend:     stateStoreBackend,
		StateStorePath:        stateStorePath,
		AuthValidationBackend: strings.ToLower(stringFromEnv("AUTH_VALIDATION_BACKEND", defaultAuthValidationBackend)),

		ServerHost:              stringFromEnv("SERVER_HOST", defaultServerHost),
		ServerPort:              serverPort,
		ReadHeaderTimeout:       time.Duration(readHeaderSec) * time.Second,
		IdleTimeout:             time.Duration(idleTimeoutSec) * time.Second,
		ShutdownTimeout:         time.Duration(shutdownTimeoutSec) * time.Second,
		AuthServiceURL:          stringFromEnv("AUTH_SERVICE_URL", defaultAuthServiceURL),
		AuthTimeout:             time.Duration(authTimeoutSec) * time.Second,
		CatalogServiceURL:       stringFromEnv("CATALOG_SERVICE_URL", defaultCatalogServiceURL),
		CatalogTimeout:          time.Duration(catalogTimeoutSec) * time.Second,
		EnableCatalogValidation: enableCatalogValidation,
		JWTActiveKeyID:          stringFromEnv("AUTH_JWT_ACTIVE_KID", defaultJWTActiveKeyID),
		JWTActiveKeySecret:      strings.TrimSpace(stringFromEnv("AUTH_JWT_ACTIVE_SECRET", "")),
		JWTPreviousKeys:         strings.TrimSpace(os.Getenv("AUTH_JWT_PREVIOUS_KEYS")),
	}

	if err := cfg.ValidateRuntimePolicy(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) ValidateRuntimePolicy() error {
	if c.KafkaTransport != "kafka" && c.KafkaTransport != "inmemory" {
		return fmt.Errorf("invalid KAFKA_TRANSPORT: %s", c.KafkaTransport)
	}
	if c.StateStoreBackend != "inmemory" && c.StateStoreBackend != "redis" && c.StateStoreBackend != "postgres" {
		return fmt.Errorf("invalid STATE_STORE_BACKEND: %s", c.StateStoreBackend)
	}
	if c.StateStoreBackend != "inmemory" && strings.TrimSpace(c.StateStorePath) == "" {
		return fmt.Errorf("STATE_STORE_PATH is required when STATE_STORE_BACKEND=%s", c.StateStoreBackend)
	}
	if c.AuthValidationBackend != "http" && c.AuthValidationBackend != "inmemory" && c.AuthValidationBackend != "jwt" {
		return fmt.Errorf("invalid AUTH_VALIDATION_BACKEND: %s", c.AuthValidationBackend)
	}
	if strings.ToLower(strings.TrimSpace(c.RuntimeProfile)) != "test" && c.StateStoreBackend == "inmemory" {
		return fmt.Errorf("STATE_STORE_BACKEND=inmemory is allowed only in test profile")
	}

	if isStrictRuntimeProfile(c.RuntimeProfile) {
		if c.KafkaTransport != "kafka" {
			return fmt.Errorf("KAFKA_TRANSPORT=inmemory is allowed only in test profile")
		}
		if strings.TrimSpace(c.KafkaBootstrapServers) == "" {
			return fmt.Errorf("KAFKA_BOOTSTRAP_SERVERS is required for non-test profiles")
		}
		if c.AuthValidationBackend != "http" && c.AuthValidationBackend != "jwt" {
			return fmt.Errorf("AUTH_VALIDATION_BACKEND must be http or jwt for non-test profiles")
		}
		if !c.EnableCatalogValidation {
			return fmt.Errorf("ENABLE_CATALOG_VALIDATION must be true for non-test profiles")
		}
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
