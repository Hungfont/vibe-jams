package gateway

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"video-streaming/backend/api-gateway/internal/config"
)

// NewRouter builds the api-gateway HTTP handler.
func NewRouter(cfg config.Config) (http.Handler, error) {
	authURL, err := url.Parse(strings.TrimRight(cfg.AuthServiceURL, "/"))
	if err != nil {
		return nil, err
	}
	apiURL, err := url.Parse(strings.TrimRight(cfg.APIServiceURL, "/"))
	if err != nil {
		return nil, err
	}

	authProxy := httputil.NewSingleHostReverseProxy(authURL)
	apiProxy := httputil.NewSingleHostReverseProxy(apiURL)
	authn := newAuthnMiddleware(cfg.AuthServiceURL, cfg.AuthTimeout)
	upstreamTimeout := cfg.UpstreamTimeout

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, _ := json.Marshal(map[string]string{"status": "ok", "service": "api-gateway"})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(gatewaySwaggerUIHTML))
	})
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/swagger/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := marshalGatewayOpenAPISpec()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), upstreamTimeout)
		defer cancel()
		r = r.WithContext(ctx)

		if !authn.apply(w, r) {
			return
		}

		// Route public auth paths to auth-service, everything else to api-service.
		if strings.HasPrefix(r.URL.Path, "/v1/auth/") {
			authProxy.ServeHTTP(w, r)
			return
		}
		apiProxy.ServeHTTP(w, r)
	})

	return requestLoggingMiddleware("api-gateway", mux), nil
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
