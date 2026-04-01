package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	defaultServerAddr       = ":8081"
	defaultRuntimeProfile   = "local"
	defaultValidatorBackend = "inmemory"
)

// Config contains runtime settings for auth-service.
type Config struct {
	ServerAddr       string
	RuntimeProfile   string
	ValidatorBackend string
}

// Load reads runtime configuration from environment.
func Load() (Config, error) {
	cfg := Config{
		ServerAddr:       stringFromEnv("SERVER_ADDR", defaultServerAddr),
		RuntimeProfile:   strings.ToLower(stringFromEnv("APP_ENV", defaultRuntimeProfile)),
		ValidatorBackend: strings.ToLower(stringFromEnv("AUTH_VALIDATOR_BACKEND", defaultValidatorBackend)),
	}
	if err := cfg.ValidateRuntimePolicy(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// ValidateRuntimePolicy validates runtime mode restrictions.
func (c Config) ValidateRuntimePolicy() error {
	if c.ValidatorBackend != "inmemory" {
		return fmt.Errorf("AUTH_VALIDATOR_BACKEND=%s is not yet implemented", c.ValidatorBackend)
	}
	if isStrictRuntimeProfile(c.RuntimeProfile) && c.ValidatorBackend == "inmemory" {
		return fmt.Errorf("AUTH_VALIDATOR_BACKEND=inmemory is not allowed in strict runtime profiles")
	}
	if strings.TrimSpace(c.ServerAddr) == "" {
		return fmt.Errorf("SERVER_ADDR is required")
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

func isStrictRuntimeProfile(profile string) bool {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "prod", "production", "staging":
		return true
	default:
		return false
	}
}
