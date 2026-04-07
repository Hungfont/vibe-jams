package bff

import (
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

	mux := http.NewServeMux()
	mux.Handle("/v1/bff/mvp/sessions/", handler)
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

	return mux, nil
}
