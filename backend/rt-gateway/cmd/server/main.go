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

	gatewaykafka "video-streaming/backend/rt-gateway/internal/kafka"
	"video-streaming/backend/rt-gateway/internal/server"
)

func main() {
	cfg, err := server.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	consumer, err := buildConsumer(cfg)
	if err != nil {
		slog.Error("failed to build consumer", "error", err)
		os.Exit(1)
	}

	app := server.NewApp(cfg, consumer)
	httpServer := &http.Server{
		Addr:              net.JoinHostPort(cfg.ServerHost, fmt.Sprintf("%d", cfg.ServerPort)),
		Handler:           app.Handler(),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	consumerErrCh := make(chan error, 1)
	go func() {
		if runErr := app.StartConsumer(ctx); runErr != nil && !errors.Is(runErr, context.Canceled) {
			consumerErrCh <- runErr
		}
	}()

	serverErrCh := make(chan error, 1)
	go func() {
		slog.Info("starting rt-gateway", "addr", httpServer.Addr, "consumerGroup", cfg.ConsumerGroupID)
		if runErr := httpServer.ListenAndServe(); runErr != nil && !errors.Is(runErr, http.ErrServerClosed) {
			serverErrCh <- fmt.Errorf("listen and serve: %w", runErr)
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case runErr := <-serverErrCh:
		slog.Error("http server exited unexpectedly", "error", runErr)
		os.Exit(1)
	case runErr := <-consumerErrCh:
		slog.Error("consumer loop exited unexpectedly", "error", runErr)
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("rt-gateway shut down cleanly")
}

func buildConsumer(cfg server.Config) (gatewaykafka.Consumer, error) {
	switch cfg.ConsumerBackend {
	case "kafka":
		topics := []string{cfg.QueueTopic, cfg.PlaybackTopic}
		consumer, err := gatewaykafka.NewKafkaConsumer(cfg.KafkaBootstrap, cfg.ConsumerGroupID, topics)
		if err != nil {
			return nil, err
		}
		return consumer, nil

	case "in-memory":
		return gatewaykafka.NewInMemoryConsumer(3000), nil

	default:
		return gatewaykafka.NewNoopConsumer(), nil
	}
}
