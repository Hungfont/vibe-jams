package bff

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"video-streaming/backend/api-service/internal/config"
)

// ProxyHandler forwards non-auth BFF route families to downstream services.
type ProxyHandler struct {
	jamProxy       *httputil.ReverseProxy
	playbackProxy  *httputil.ReverseProxy
	catalogProxy   *httputil.ReverseProxy
	rtGatewayProxy *httputil.ReverseProxy
	gatewayWSBase  string
}

// NewProxyHandler builds route-family proxies for jam, playback, catalog, and realtime bootstrap config.
func NewProxyHandler(cfg config.Config) (*ProxyHandler, error) {
	jamURL, err := url.Parse(strings.TrimRight(cfg.JamServiceURL, "/"))
	if err != nil {
		return nil, err
	}
	playbackURL, err := url.Parse(strings.TrimRight(cfg.PlaybackServiceURL, "/"))
	if err != nil {
		return nil, err
	}
	catalogURL, err := url.Parse(strings.TrimRight(cfg.CatalogServiceURL, "/"))
	if err != nil {
		return nil, err
	}
	rtGatewayURL, err := url.Parse(strings.TrimRight(cfg.RTGatewayURL, "/"))
	if err != nil {
		return nil, err
	}

	return &ProxyHandler{
		jamProxy:       httputil.NewSingleHostReverseProxy(jamURL),
		playbackProxy:  httputil.NewSingleHostReverseProxy(playbackURL),
		catalogProxy:   httputil.NewSingleHostReverseProxy(catalogURL),
		rtGatewayProxy: newRealtimeWSProxy(rtGatewayURL),
		gatewayWSBase:  toWSURL(cfg.GatewayPublicURL) + "/v1/bff/mvp/realtime/ws",
	}, nil
}

// ServeHTTP routes path families to downstream services through api-service.
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/v1/jams/"):
		h.jamProxy.ServeHTTP(w, r)
		return
	case strings.HasPrefix(r.URL.Path, "/v1/jam/sessions/"):
		h.playbackProxy.ServeHTTP(w, r)
		return
	case strings.HasPrefix(r.URL.Path, "/internal/v1/catalog/tracks/"):
		h.catalogProxy.ServeHTTP(w, r)
		return
	case r.URL.Path == "/v1/bff/mvp/realtime/ws":
		h.rtGatewayProxy.ServeHTTP(w, r)
		return
	case r.URL.Path == "/v1/bff/mvp/realtime/ws-config":
		h.handleRealtimeWSConfig(w, r)
		return
	default:
		http.NotFound(w, r)
		return
	}
}

func (h *ProxyHandler) handleRealtimeWSConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := strings.TrimSpace(r.URL.Query().Get("sessionId"))
	if sessionID == "" {
		http.Error(w, "sessionId is required", http.StatusBadRequest)
		return
	}
	lastSeenVersion := strings.TrimSpace(r.URL.Query().Get("lastSeenVersion"))

	payload := map[string]string{
		"wsUrl":           h.gatewayWSBase,
		"sessionId":       sessionID,
		"lastSeenVersion": lastSeenVersion,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func toWSURL(base string) string {
	if strings.HasPrefix(base, "https://") {
		return "wss://" + strings.TrimPrefix(base, "https://")
	}
	if strings.HasPrefix(base, "http://") {
		return "ws://" + strings.TrimPrefix(base, "http://")
	}
	return base
}

func newRealtimeWSProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.URL.Path = "/ws"
	}
	return proxy
}
