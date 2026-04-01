package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"video-streaming/backend/jams/internal/config"
	"video-streaming/backend/jams/internal/server"
)

// main boots the HTTP server and handles graceful shutdown.
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	router, err := server.NewRouter(cfg)
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

	serverErrCh := make(chan error, 1)
	go func() {
		slog.Info("starting http server", "addr", addr)
		if listenErr := httpServer.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			serverErrCh <- fmt.Errorf("listen and serve: %w", listenErr)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case runErr := <-serverErrCh:
		slog.Error("server stopped unexpectedly", "error", runErr)
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server shut down cleanly")
}
