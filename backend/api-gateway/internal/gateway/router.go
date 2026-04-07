package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

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

	return mux, nil
}
