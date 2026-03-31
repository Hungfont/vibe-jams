package main

import (
	"log/slog"
	"net/http"
	"os"

	"video-streaming/backend/auth-service/internal/auth"
	httpserver "video-streaming/backend/auth-service/internal/http"
)

func main() {
	validator := auth.NewInMemoryValidator()
	handler := httpserver.NewHandler(validator)

	addr := ":8081"
	if envAddr := os.Getenv("SERVER_ADDR"); envAddr != "" {
		addr = envAddr
	}

	slog.Info("starting auth-service", "addr", addr)
	if err := http.ListenAndServe(addr, handler.Router()); err != nil {
		slog.Error("auth-service stopped", "error", err)
		os.Exit(1)
	}
}
