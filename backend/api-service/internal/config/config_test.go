package config

import "testing"

func TestLoadDefaultConfig(t *testing.T) {
	t.Setenv("SERVER_PORT", "")
	t.Setenv("FEATURE_BFF_ENABLED", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ServerPort != defaultServerPort {
		t.Fatalf("ServerPort mismatch: got %d want %d", cfg.ServerPort, defaultServerPort)
	}
	if !cfg.FeatureBFFEnabled {
		t.Fatal("expected FEATURE_BFF_ENABLED default true")
	}
}

func TestLoadRejectsInvalidTimeout(t *testing.T) {
	t.Setenv("BFF_TIMEOUT_JAM_MS", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for invalid timeout")
	}
}
