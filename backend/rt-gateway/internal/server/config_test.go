package server

import "testing"

func TestLoadConfigRejectsNoopInNonTestProfile(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("KAFKA_CONSUMER_BACKEND", "noop")
	t.Setenv("KAFKA_BOOTSTRAP_SERVERS", "kafka:9092")
	t.Setenv("WS_ALLOWED_ORIGINS", "https://app.example.com")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for noop consumer in non-test profile")
	}
}

func TestLoadConfigRequiresOriginAllowlistInNonTestProfile(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("KAFKA_CONSUMER_BACKEND", "kafka")
	t.Setenv("KAFKA_BOOTSTRAP_SERVERS", "kafka:9092")
	t.Setenv("WS_ALLOWED_ORIGINS", "")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error when WS_ALLOWED_ORIGINS is empty in non-test profile")
	}
}

func TestLoadConfigLocalFallbacks(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("KAFKA_CONSUMER_BACKEND", "kafka")
	t.Setenv("KAFKA_BOOTSTRAP_SERVERS", "")
	t.Setenv("WS_ALLOWED_ORIGINS", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.KafkaBootstrap == "" {
		t.Fatal("expected localhost Kafka bootstrap fallback in local profile")
	}
	if len(cfg.AllowedOrigins) == 0 {
		t.Fatal("expected default allowed origins in local profile")
	}
}
