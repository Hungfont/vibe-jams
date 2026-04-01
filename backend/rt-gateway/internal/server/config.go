package server

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultServerHost           = "0.0.0.0"
	defaultServerPort           = 8090
	defaultReadHeaderTimeoutSec = 5
	defaultIdleTimeoutSec       = 60
	defaultShutdownTimeoutSec   = 10
	defaultJamServiceURL        = "http://localhost:8080"
	defaultSnapshotTimeoutSec   = 2
	defaultFanoutBufferSize     = 64
	defaultConsumerGroupID      = "rt-gateway-fanout"
	defaultQueueTopic           = "jam.queue.events"
	defaultPlaybackTopic        = "jam.playback.events"
	defaultRecoveryMaxRetries   = 3
	defaultRecoveryBackoffMS    = 200
	defaultFeatureFanoutEnabled = true
	defaultRuntimeProfile       = "local"
	defaultKafkaBootstrap       = "localhost:9092"
	defaultConsumerBackend      = "in-memory"
	defaultOriginAllowlist      = "http://localhost:3000,http://127.0.0.1:3000"
)

// Config contains runtime settings for realtime fanout.
type Config struct {
	RuntimeProfile string

	ServerHost string
	ServerPort int

	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration

	JamServiceURL    string
	SnapshotTimeout  time.Duration
	FanoutBufferSize int

	ConsumerGroupID string
	KafkaBootstrap  string
	ConsumerBackend string
	QueueTopic      string
	PlaybackTopic   string
	AllowedOrigins  []string

	RecoveryMaxRetries int
	RecoveryBackoff    time.Duration

	FeatureFanoutEnabled bool
}

// LoadConfig resolves environment variables into process configuration.
func LoadConfig() (Config, error) {
	serverPort, err := intFromEnv("SERVER_PORT", defaultServerPort)
	if err != nil {
		return Config{}, fmt.Errorf("parse SERVER_PORT: %w", err)
	}

	readHeaderSec, err := intFromEnv("READ_HEADER_TIMEOUT_SEC", defaultReadHeaderTimeoutSec)
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

	snapshotTimeoutSec, err := intFromEnv("SNAPSHOT_TIMEOUT_SEC", defaultSnapshotTimeoutSec)
	if err != nil {
		return Config{}, fmt.Errorf("parse SNAPSHOT_TIMEOUT_SEC: %w", err)
	}

	fanoutBufferSize, err := intFromEnv("FANOUT_BUFFER_SIZE", defaultFanoutBufferSize)
	if err != nil {
		return Config{}, fmt.Errorf("parse FANOUT_BUFFER_SIZE: %w", err)
	}
	if fanoutBufferSize <= 0 {
		return Config{}, fmt.Errorf("invalid FANOUT_BUFFER_SIZE: must be > 0")
	}

	recoveryMaxRetries, err := intFromEnv("RECOVERY_MAX_RETRIES", defaultRecoveryMaxRetries)
	if err != nil {
		return Config{}, fmt.Errorf("parse RECOVERY_MAX_RETRIES: %w", err)
	}
	if recoveryMaxRetries < 0 {
		return Config{}, fmt.Errorf("invalid RECOVERY_MAX_RETRIES: must be >= 0")
	}

	recoveryBackoffMS, err := intFromEnv("RECOVERY_BACKOFF_MS", defaultRecoveryBackoffMS)
	if err != nil {
		return Config{}, fmt.Errorf("parse RECOVERY_BACKOFF_MS: %w", err)
	}
	if recoveryBackoffMS <= 0 {
		return Config{}, fmt.Errorf("invalid RECOVERY_BACKOFF_MS: must be > 0")
	}

	featureFanoutEnabled, err := boolFromEnv("FEATURE_REALTIME_FANOUT_ENABLED", defaultFeatureFanoutEnabled)
	if err != nil {
		return Config{}, fmt.Errorf("parse FEATURE_REALTIME_FANOUT_ENABLED: %w", err)
	}

	runtimeProfile := strings.ToLower(stringFromEnv("APP_ENV", defaultRuntimeProfile))
	kafkaBootstrap := strings.TrimSpace(os.Getenv("KAFKA_BOOTSTRAP_SERVERS"))
	if kafkaBootstrap == "" && runtimeProfile == "local" {
		kafkaBootstrap = defaultKafkaBootstrap
	}

	consumerBackend := strings.ToLower(stringFromEnv("KAFKA_CONSUMER_BACKEND", defaultConsumerBackend))
	originAllowlistRaw := strings.TrimSpace(os.Getenv("WS_ALLOWED_ORIGINS"))
	if originAllowlistRaw == "" && runtimeProfile == "local" {
		originAllowlistRaw = defaultOriginAllowlist
	}
	originAllowlist := parseCSV(originAllowlistRaw)

	cfg := Config{
		RuntimeProfile: runtimeProfile,

		ServerHost: stringFromEnv("SERVER_HOST", defaultServerHost),
		ServerPort: serverPort,

		ReadHeaderTimeout: time.Duration(readHeaderSec) * time.Second,
		IdleTimeout:       time.Duration(idleSec) * time.Second,
		ShutdownTimeout:   time.Duration(shutdownSec) * time.Second,

		JamServiceURL:    stringFromEnv("JAM_SERVICE_URL", defaultJamServiceURL),
		SnapshotTimeout:  time.Duration(snapshotTimeoutSec) * time.Second,
		FanoutBufferSize: fanoutBufferSize,

		ConsumerGroupID: stringFromEnv("KAFKA_CONSUMER_GROUP", defaultConsumerGroupID),
		KafkaBootstrap:  kafkaBootstrap,
		ConsumerBackend: consumerBackend,
		QueueTopic:      stringFromEnv("KAFKA_TOPIC_QUEUE", defaultQueueTopic),
		PlaybackTopic:   stringFromEnv("KAFKA_TOPIC_PLAYBACK", defaultPlaybackTopic),
		AllowedOrigins:  originAllowlist,

		RecoveryMaxRetries: recoveryMaxRetries,
		RecoveryBackoff:    time.Duration(recoveryBackoffMS) * time.Millisecond,

		FeatureFanoutEnabled: featureFanoutEnabled,
	}

	if err := cfg.validateRuntimePolicy(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validateRuntimePolicy() error {
	if c.ConsumerBackend != "kafka" && c.ConsumerBackend != "noop" {
		return fmt.Errorf("invalid KAFKA_CONSUMER_BACKEND: %s", c.ConsumerBackend)
	}

	if c.RuntimeProfile != "test" && c.ConsumerBackend == "noop" {
		return fmt.Errorf("KAFKA_CONSUMER_BACKEND=noop is allowed only in test profile")
	}

	if c.RuntimeProfile != "test" && strings.TrimSpace(c.KafkaBootstrap) == "" {
		return fmt.Errorf("KAFKA_BOOTSTRAP_SERVERS is required for non-test profiles")
	}

	if c.RuntimeProfile != "test" && len(c.AllowedOrigins) == 0 {
		return fmt.Errorf("WS_ALLOWED_ORIGINS is required for non-test profiles")
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

func parseCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}
