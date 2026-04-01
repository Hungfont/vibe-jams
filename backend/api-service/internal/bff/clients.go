package bff

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	sharedauth "video-streaming/backend/shared/auth"
	sharedcatalog "video-streaming/backend/shared/catalog"
)

// AuthClient validates bearer token and returns normalized claims.
type AuthClient interface {
	ValidateBearerToken(ctx context.Context, authHeader string) (sharedauth.Claims, error)
}

// JamClient loads jam session state.
type JamClient interface {
	SessionState(ctx context.Context, jamID string, authHeader string) (SessionStateSnapshot, error)
}

// PlaybackClient sends playback commands.
type PlaybackClient interface {
	ExecuteCommand(ctx context.Context, jamID string, authHeader string, req PlaybackCommandRequest) (PlaybackCommandAccepted, error)
}

// CatalogClient validates track metadata.
type CatalogClient interface {
	LookupTrack(ctx context.Context, trackID string) (LookupResponse, error)
}

// HTTPAuthClient calls auth-service.
type HTTPAuthClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPAuthClient builds auth dependency client.
func NewHTTPAuthClient(baseURL string, timeout time.Duration) *HTTPAuthClient {
	return &HTTPAuthClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: timeout}}
}

func (c *HTTPAuthClient) ValidateBearerToken(ctx context.Context, authHeader string) (sharedauth.Claims, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/auth/validate", bytes.NewReader(nil))
	if err != nil {
		return sharedauth.Claims{}, fmt.Errorf("build auth request: %w", err)
	}
	req.Header.Set("Authorization", strings.TrimSpace(authHeader))

	resp, err := c.client.Do(req)
	if err != nil {
		return sharedauth.Claims{}, fmt.Errorf("%w: %v", classifyTransportError(err), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return sharedauth.Claims{}, ErrUnauthorized
	}
	if resp.StatusCode >= http.StatusInternalServerError {
		return sharedauth.Claims{}, ErrDependencyUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		return sharedauth.Claims{}, UpstreamError{StatusCode: resp.StatusCode, Code: "upstream_error", Message: "unexpected auth status"}
	}

	var claims sharedauth.Claims
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return sharedauth.Claims{}, fmt.Errorf("decode claims: %w", err)
	}
	if err := sharedauth.ValidateClaims(claims); err != nil {
		return sharedauth.Claims{}, ErrUnauthorized
	}
	if strings.ToLower(strings.TrimSpace(claims.SessionState)) != sharedauth.SessionStateValid {
		return sharedauth.Claims{}, ErrUnauthorized
	}
	return claims, nil
}

// HTTPJamClient calls jam-service state endpoint.
type HTTPJamClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPJamClient builds jam dependency client.
func NewHTTPJamClient(baseURL string, timeout time.Duration) *HTTPJamClient {
	return &HTTPJamClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: timeout}}
}

func (c *HTTPJamClient) SessionState(ctx context.Context, jamID string, authHeader string) (SessionStateSnapshot, error) {
	endpoint := fmt.Sprintf("%s/api/v1/jams/%s/state", c.baseURL, url.PathEscape(jamID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return SessionStateSnapshot{}, fmt.Errorf("build jam state request: %w", err)
	}
	if strings.TrimSpace(authHeader) != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return SessionStateSnapshot{}, fmt.Errorf("%w: %v", classifyTransportError(err), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return SessionStateSnapshot{}, ErrNotFound
	}
	if resp.StatusCode >= http.StatusInternalServerError {
		return SessionStateSnapshot{}, ErrDependencyUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		var up struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&up)
		return SessionStateSnapshot{}, UpstreamError{StatusCode: resp.StatusCode, Code: up.Error.Code, Message: up.Error.Message}
	}

	var out SessionStateSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return SessionStateSnapshot{}, fmt.Errorf("decode jam state: %w", err)
	}
	return out, nil
}

// HTTPPlaybackClient calls playback-service command endpoint.
type HTTPPlaybackClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPPlaybackClient builds playback dependency client.
func NewHTTPPlaybackClient(baseURL string, timeout time.Duration) *HTTPPlaybackClient {
	return &HTTPPlaybackClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: timeout}}
}

func (c *HTTPPlaybackClient) ExecuteCommand(ctx context.Context, jamID string, authHeader string, reqBody PlaybackCommandRequest) (PlaybackCommandAccepted, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return PlaybackCommandAccepted{}, fmt.Errorf("marshal playback request: %w", err)
	}
	endpoint := fmt.Sprintf("%s/v1/jam/sessions/%s/playback/commands", c.baseURL, url.PathEscape(jamID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return PlaybackCommandAccepted{}, fmt.Errorf("build playback request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(authHeader) != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return PlaybackCommandAccepted{}, fmt.Errorf("%w: %v", classifyTransportError(err), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		return PlaybackCommandAccepted{}, ErrDependencyUnavailable
	}
	if resp.StatusCode != http.StatusAccepted {
		var up struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&up)
		return PlaybackCommandAccepted{}, UpstreamError{StatusCode: resp.StatusCode, Code: up.Error.Code, Message: up.Error.Message}
	}

	var out PlaybackCommandAccepted
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return PlaybackCommandAccepted{}, fmt.Errorf("decode playback response: %w", err)
	}
	return out, nil
}

// HTTPCatalogClient validates tracks through shared contract client.
type HTTPCatalogClient struct {
	validator sharedcatalog.Validator
}

// NewHTTPCatalogClient builds catalog dependency client.
func NewHTTPCatalogClient(baseURL string, timeout time.Duration) *HTTPCatalogClient {
	return &HTTPCatalogClient{validator: sharedcatalog.NewHTTPValidator(strings.TrimRight(baseURL, "/"), timeout)}
}

func (c *HTTPCatalogClient) LookupTrack(ctx context.Context, trackID string) (LookupResponse, error) {
	result, err := c.validator.ValidateTrack(ctx, strings.TrimSpace(trackID))
	if err != nil {
		if errors.Is(err, sharedcatalog.ErrDependencyUnavailable) {
			return LookupResponse{}, ErrDependencyUnavailable
		}
		if errors.Is(err, sharedcatalog.ErrTrackNotFound) {
			return LookupResponse{}, UpstreamError{StatusCode: http.StatusNotFound, Code: "track_not_found", Message: "track not found"}
		}
		if errors.Is(err, sharedcatalog.ErrTrackUnavailable) {
			return LookupResponse{TrackID: result.TrackID, IsPlayable: false, ReasonCode: result.ReasonCode, Title: result.Title, Artist: result.Artist}, UpstreamError{StatusCode: http.StatusConflict, Code: "track_unavailable", Message: "track unavailable"}
		}
		return LookupResponse{}, err
	}
	return LookupResponse{TrackID: result.TrackID, IsPlayable: result.IsPlayable, ReasonCode: result.ReasonCode, Title: result.Title, Artist: result.Artist}, nil
}

func classifyTransportError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrDependencyTimeout
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return ErrDependencyTimeout
	}
	return ErrDependencyUnavailable
}
