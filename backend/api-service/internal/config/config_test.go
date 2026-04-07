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
	if cfg.GatewayPublicURL != defaultGatewayPublicURL {
		t.Fatalf("GatewayPublicURL mismatch: got %q want %q", cfg.GatewayPublicURL, defaultGatewayPublicURL)
	}
}

func TestLoadRejectsInvalidTimeout(t *testing.T) {
	t.Setenv("BFF_TIMEOUT_JAM_MS", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for invalid timeout")
	}
}

func TestLoadRejectsMissingGatewayPublicURL(t *testing.T) {
	t.Setenv("GATEWAY_PUBLIC_URL", "")
	t.Setenv("JAM_SERVICE_URL", "http://localhost:8080")
	t.Setenv("PLAYBACK_SERVICE_URL", "http://localhost:8082")
	t.Setenv("CATALOG_SERVICE_URL", "http://localhost:8083")
	t.Setenv("RT_GATEWAY_URL", "http://localhost:8086")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected default gateway public URL fallback, got error: %v", err)
	}
	if cfg.GatewayPublicURL == "" {
		t.Fatal("GatewayPublicURL must not be empty")
	}
}
