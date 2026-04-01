package config

import "testing"

func TestLoadRejectsInMemoryBackendInStrictProfiles(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("AUTH_VALIDATOR_BACKEND", "inmemory")

	_, err := Load()
	if err == nil {
		t.Fatal("expected strict profile to reject in-memory validator backend")
	}
}

func TestLoadAllowsInMemoryBackendInLocalProfile(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("AUTH_VALIDATOR_BACKEND", "inmemory")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ValidatorBackend != "inmemory" {
		t.Fatalf("backend mismatch: got %q", cfg.ValidatorBackend)
	}
}
