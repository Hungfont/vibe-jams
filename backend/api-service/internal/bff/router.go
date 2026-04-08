package bff

import (
	"bufio"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"video-streaming/backend/api-service/internal/config"
)

// NewRouter builds all api-service routes.
func NewRouter(cfg config.Config) (http.Handler, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	jamClient := NewHTTPJamClient(cfg.JamServiceURL, cfg.JamTimeout)
	playbackClient := NewHTTPPlaybackClient(cfg.PlaybackServiceURL, cfg.PlaybackTimeout)
	catalogClient := NewHTTPCatalogClient(cfg.CatalogServiceURL, cfg.CatalogTimeout)

	service := NewService(jamClient, playbackClient, catalogClient, cfg.FeatureBFFEnabled)
	handler := NewHandler(service)
	proxyHandler, err := NewProxyHandler(cfg)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/bff/mvp/sessions/", handler)
	mux.Handle("/api/v1/jams/", proxyHandler)
	mux.Handle("/v1/jam/sessions/", proxyHandler)
	mux.Handle("/internal/v1/catalog/tracks/", proxyHandler)
	mux.Handle("/v1/bff/mvp/realtime/ws", proxyHandler)
	mux.Handle("/v1/bff/mvp/realtime/ws-config", proxyHandler)
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(swaggerUIHTML))
	})
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":    "ok",
			"service":   "api-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})
	mux.HandleFunc("/swagger/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := marshalOpenAPISpec()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	return requestLoggingMiddleware("api-service", mux), nil
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("response writer does not support hijacking")
	}
	return hijacker.Hijack()
}

func requestLoggingMiddleware(service string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)
		slog.Info("http request",
			"service", service,
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.statusCode,
			"durationMs", time.Since(start).Milliseconds(),
		)
	})
}
