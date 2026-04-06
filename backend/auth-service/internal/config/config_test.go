package config

import "testing"

func TestLoadRejectsInMemoryBackendInStrictProfiles(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_SESSION_STORE_BACKEND", "inmemory")
	t.Setenv("AUTH_JWT_ACTIVE_SECRET", "test-secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected production profile to reject in-memory session backend")
	}
}

func TestLoadAllowsInMemoryBackendInLocalProfile(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("AUTH_SESSION_STORE_BACKEND", "inmemory")
	t.Setenv("AUTH_JWT_ACTIVE_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.SessionStoreBackend != "inmemory" {
		t.Fatalf("backend mismatch: got %q", cfg.SessionStoreBackend)
	}
}

func TestLoadRejectsPostgresBackendWithoutDSN(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("AUTH_SESSION_STORE_BACKEND", "postgres")
	t.Setenv("AUTH_JWT_ACTIVE_SECRET", "test-secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected postgres backend to require DSN")
	}
}
