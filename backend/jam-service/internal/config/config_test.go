package config

import "testing"

func TestLoadRejectsInMemoryKafkaInNonTestProfile(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("KAFKA_TRANSPORT", "inmemory")
	t.Setenv("KAFKA_BOOTSTRAP_SERVERS", "kafka:9092")
	t.Setenv("ENABLE_CATALOG_VALIDATION", "true")
	t.Setenv("AUTH_VALIDATION_BACKEND", "http")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for KAFKA_TRANSPORT=inmemory in non-test profile")
	}
}

func TestLoadRequiresCatalogValidationInNonTestProfile(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("KAFKA_TRANSPORT", "kafka")
	t.Setenv("KAFKA_BOOTSTRAP_SERVERS", "kafka:9092")
	t.Setenv("ENABLE_CATALOG_VALIDATION", "false")
	t.Setenv("AUTH_VALIDATION_BACKEND", "http")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for disabled catalog validation in non-test profile")
	}
}

func TestLoadLocalKafkaFallback(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("KAFKA_TRANSPORT", "kafka")
	t.Setenv("KAFKA_BOOTSTRAP_SERVERS", "")
	t.Setenv("ENABLE_CATALOG_VALIDATION", "true")
	t.Setenv("AUTH_VALIDATION_BACKEND", "http")
	t.Setenv("STATE_STORE_BACKEND", "redis")
	t.Setenv("STATE_STORE_PATH", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.KafkaBootstrapServers == "" {
		t.Fatal("expected local Kafka bootstrap fallback")
	}
	if cfg.StateStorePath == "" {
		t.Fatal("expected local state store path fallback")
	}
}

func TestLoadRejectsInMemoryStateStoreInNonTestProfile(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("KAFKA_TRANSPORT", "kafka")
	t.Setenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
	t.Setenv("ENABLE_CATALOG_VALIDATION", "true")
	t.Setenv("AUTH_VALIDATION_BACKEND", "http")
	t.Setenv("STATE_STORE_BACKEND", "inmemory")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for STATE_STORE_BACKEND=inmemory in non-test profile")
	}
}
