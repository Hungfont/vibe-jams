package bff

import (
	"bytes"
	"encoding/json"
	"io"
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
	jamStateClient JamClient
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
		jamStateClient: NewHTTPJamClient(cfg.JamServiceURL, cfg.JamTimeout),
	}, nil
}

// ServeHTTP routes path families to downstream services through api-service.
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/v1/jams/"):
		if jamID, needsHostOnly := delegatedPolicyRoute(r.URL.Path, r.Method); needsHostOnly {
			if !h.authorizeHostOnlyPolicyCommand(w, r, jamID) {
				return
			}
		}
		if jamID, command, needsPermission := delegatedProtectedQueueRoute(r.URL.Path, r.Method); needsPermission {
			if !h.authorizeProjectedPermission(w, r, jamID, command) {
				return
			}
		}
		h.jamProxy.ServeHTTP(w, r)
		return
	case strings.HasPrefix(r.URL.Path, "/v1/jam/sessions/"):
		if jamID, command, needsPermission, ok := delegatedPlaybackRoute(r); ok {
			if needsPermission {
				if !h.authorizeProjectedPermission(w, r, jamID, command) {
					return
				}
			}
		}
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

func (h *ProxyHandler) authorizeHostOnlyPolicyCommand(w http.ResponseWriter, r *http.Request, jamID string) bool {
	claims, xAuthHeaders, ok := extractGatewayIdentity(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, Envelope{Success: false, Error: &ErrorBody{Code: "unauthorized", Message: "missing or invalid identity context"}})
		return false
	}

	state, err := h.jamStateClient.SessionState(r.Context(), jamID, xAuthHeaders)
	if err != nil {
		errBody, status := mapPolicyAuthzDependencyError(err)
		writeJSON(w, status, Envelope{Success: false, Error: errBody})
		return false
	}

	hostUserID := strings.TrimSpace(state.Session.HostUserID)
	if hostUserID == "" || claims.UserID != hostUserID {
		writeJSON(w, http.StatusForbidden, Envelope{Success: false, Error: &ErrorBody{Code: "host_only", Message: "host-only policy command"}})
		return false
	}

	return true
}

func mapPolicyAuthzDependencyError(err error) (*ErrorBody, int) {
	code, message := mapDependencyError(err)
	status := http.StatusServiceUnavailable
	switch code {
	case "unauthorized":
		status = http.StatusUnauthorized
	case "not_found":
		status = http.StatusNotFound
	case "dependency_timeout", "dependency_unavailable":
		status = http.StatusServiceUnavailable
	default:
		status = http.StatusServiceUnavailable
	}

	return &ErrorBody{Code: code, Message: message, Dependency: "jam"}, status
}

func (h *ProxyHandler) authorizeProjectedPermission(w http.ResponseWriter, r *http.Request, jamID string, command projectedPermissionCommand) bool {
	claims, xAuthHeaders, ok := extractGatewayIdentity(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, Envelope{Success: false, Error: &ErrorBody{Code: "unauthorized", Message: "missing or invalid identity context"}})
		return false
	}

	state, err := h.jamStateClient.SessionState(r.Context(), jamID, xAuthHeaders)
	if err != nil {
		errBody, status := mapPolicyAuthzDependencyError(err)
		writeJSON(w, status, Envelope{Success: false, Error: errBody})
		return false
	}

	if strings.TrimSpace(state.Session.HostUserID) == claims.UserID {
		return true
	}

	allowed := false
	switch command {
	case commandPermissionPlayback:
		allowed = state.Session.Permissions.CanControlPlayback
	case commandPermissionReorder:
		allowed = state.Session.Permissions.CanReorderQueue
	case commandPermissionVolume:
		allowed = state.Session.Permissions.CanChangeVolume
	default:
		allowed = false
	}

	if !allowed {
		writeJSON(w, http.StatusForbidden, Envelope{Success: false, Error: &ErrorBody{Code: "permission_denied", Message: "insufficient permission"}})
		return false
	}

	return true
}

func delegatedPolicyRoute(path string, method string) (jamID string, needsHostOnly bool) {
	const prefix = "/api/v1/jams/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}

	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	parts := strings.Split(rest, "/")
	if len(parts) < 3 {
		return "", false
	}

	jamID = strings.TrimSpace(parts[0])
	if jamID == "" {
		return "", false
	}

	section := strings.ToLower(strings.TrimSpace(parts[1]))
	action := strings.ToLower(strings.TrimSpace(parts[2]))

	if section == "moderation" && (action == "mute" || action == "kick") {
		return jamID, true
	}
	if section == "permission" {
		return jamID, true
	}
	if section == "permissions" && strings.ToUpper(strings.TrimSpace(method)) != http.MethodGet {
		return jamID, true
	}

	return "", false
}

type projectedPermissionCommand string

const (
	commandPermissionPlayback projectedPermissionCommand = "permission.control_playback"
	commandPermissionReorder  projectedPermissionCommand = "permission.reorder_queue"
	commandPermissionVolume   projectedPermissionCommand = "permission.change_volume"
)

func delegatedProtectedQueueRoute(path string, method string) (jamID string, command projectedPermissionCommand, needsPermission bool) {
	if strings.ToUpper(strings.TrimSpace(method)) != http.MethodPost {
		return "", "", false
	}
	const prefix = "/api/v1/jams/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	parts := strings.Split(rest, "/")
	if len(parts) != 3 || strings.TrimSpace(parts[0]) == "" {
		return "", "", false
	}
	if parts[1] == "queue" && parts[2] == "reorder" {
		return strings.TrimSpace(parts[0]), commandPermissionReorder, true
	}
	return "", "", false
}

func delegatedPlaybackRoute(r *http.Request) (jamID string, command projectedPermissionCommand, needsPermission bool, ok bool) {
	if strings.ToUpper(strings.TrimSpace(r.Method)) != http.MethodPost {
		return "", "", false, false
	}
	const prefix = "/v1/jam/sessions/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		return "", "", false, false
	}
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, prefix), "/")
	parts := strings.Split(rest, "/")
	if len(parts) != 3 || parts[1] != "playback" || parts[2] != "commands" {
		return "", "", false, false
	}
	jamID = strings.TrimSpace(parts[0])
	if jamID == "" {
		return "", "", false, false
	}

	commandName, err := extractPlaybackCommand(r)
	if err != nil {
		return "", "", false, false
	}
	if strings.EqualFold(commandName, "volume") {
		return jamID, commandPermissionVolume, true, true
	}
	return jamID, commandPermissionPlayback, true, true
}

func extractPlaybackCommand(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	if len(body) == 0 {
		return "", nil
	}
	var payload struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	return strings.TrimSpace(payload.Command), nil
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
