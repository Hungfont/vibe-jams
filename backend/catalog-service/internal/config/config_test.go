package config

import "testing"

func TestLoadRejectsInMemoryBackendInStrictProfiles(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("CATALOG_SOURCE_BACKEND", "inmemory")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for in-memory backend in strict runtime profile")
	}
}

func TestLoadAllowsInMemoryBackendInLocalProfile(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("CATALOG_SOURCE_BACKEND", "inmemory")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CatalogBackend != "inmemory" {
		t.Fatalf("backend mismatch: got %q want inmemory", cfg.CatalogBackend)
	}
}
