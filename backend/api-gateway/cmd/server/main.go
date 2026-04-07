package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"video-streaming/backend/api-gateway/internal/config"
	"video-streaming/backend/api-gateway/internal/gateway"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	router, err := gateway.NewRouter(cfg)
	if err != nil {
		slog.Error("failed to build router", "error", err)
		os.Exit(1)
	}

	addr := net.JoinHostPort(cfg.ServerHost, fmt.Sprintf("%d", cfg.ServerPort))
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	slog.Info("starting api-gateway", "addr", addr)
	if err := httpServer.ListenAndServe(); err != nil {
		slog.Error("api-gateway stopped", "error", err)
		os.Exit(1)
	}
}
