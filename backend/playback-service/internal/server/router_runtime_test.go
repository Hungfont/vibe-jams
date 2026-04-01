package server

import (
	"path/filepath"
	"testing"
	"time"

	"video-streaming/backend/playback-service/internal/config"
)

func TestNewRouter_DurableStateBackend(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		RuntimeProfile:          "test",
		KafkaTransport:          "inmemory",
		StateStoreBackend:       "redis",
		StateStorePath:          filepath.Join(t.TempDir(), "playback-state.json"),
		AuthValidationBackend:   "http",
		ServerHost:              "127.0.0.1",
		ServerPort:              0,
		ReadHeaderTimeout:       time.Second,
		IdleTimeout:             time.Second,
		ShutdownTimeout:         time.Second,
		AuthServiceURL:          "http://localhost:8081",
		AuthTimeout:             time.Second,
		EnableCatalogValidation: false,
	}

	h, err := NewRouter(cfg)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil router")
	}
}
