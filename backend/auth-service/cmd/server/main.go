package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"video-streaming/backend/auth-service/internal/auth"
	"video-streaming/backend/auth-service/internal/config"
	httpserver "video-streaming/backend/auth-service/internal/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	var validator auth.Validator
	switch cfg.ValidatorBackend {
	case "inmemory":
		validator = auth.NewInMemoryValidator()
	default:
		slog.Error("failed to build validator", "error", fmt.Errorf("unsupported AUTH_VALIDATOR_BACKEND: %s", cfg.ValidatorBackend))
		os.Exit(1)
	}
	handler := httpserver.NewHandler(validator)

	slog.Info("starting auth-service", "addr", cfg.ServerAddr, "profile", cfg.RuntimeProfile)
	if err := http.ListenAndServe(cfg.ServerAddr, handler.Router()); err != nil {
		slog.Error("auth-service stopped", "error", err)
		os.Exit(1)
	}
}
