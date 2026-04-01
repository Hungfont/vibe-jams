package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// ErrTrackNotFound indicates the lookup target does not exist in catalog.
	ErrTrackNotFound = errors.New("track not found")
	// ErrTrackUnavailable indicates the track exists but cannot be played.
	ErrTrackUnavailable = errors.New("track unavailable")
	// ErrDependencyUnavailable indicates catalog integration dependency is unavailable.
	ErrDependencyUnavailable = errors.New("catalog dependency unavailable")
)

// LookupResponse defines the shared catalog track validation contract.
type LookupResponse struct {
	TrackID    string `json:"trackId"`
	IsPlayable bool   `json:"isPlayable"`
	ReasonCode string `json:"reasonCode,omitempty"`
	Title      string `json:"title,omitempty"`
	Artist     string `json:"artist,omitempty"`
}

// Validator validates a track identifier against catalog contract.
type Validator interface {
	ValidateTrack(ctx context.Context, trackID string) (LookupResponse, error)
}

// HTTPValidator validates track IDs through catalog-service HTTP API.
type HTTPValidator struct {
	baseURL string
	client  *http.Client
}

// NewHTTPValidator creates a catalog-service client validator.
func NewHTTPValidator(baseURL string, timeout time.Duration) *HTTPValidator {
	return &HTTPValidator{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// ValidateTrack requests catalog and maps outcomes to deterministic errors.
func (v *HTTPValidator) ValidateTrack(ctx context.Context, trackID string) (LookupResponse, error) {
	trackID = strings.TrimSpace(trackID)
	if trackID == "" {
		return LookupResponse{}, fmt.Errorf("trackId is required")
	}
	if v.baseURL == "" {
		return LookupResponse{}, fmt.Errorf("catalog baseURL is required")
	}

	endpoint := fmt.Sprintf("%s/internal/v1/catalog/tracks/%s", v.baseURL, url.PathEscape(trackID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return LookupResponse{}, fmt.Errorf("build catalog request: %w", err)
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return LookupResponse{}, fmt.Errorf("%w: %v", ErrDependencyUnavailable, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return LookupResponse{}, ErrTrackNotFound
	case http.StatusOK:
		var out LookupResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return LookupResponse{}, fmt.Errorf("decode lookup response: %w", err)
		}
		if strings.TrimSpace(out.TrackID) == "" {
			return LookupResponse{}, fmt.Errorf("lookup response missing trackId")
		}
		if !out.IsPlayable {
			if strings.TrimSpace(out.ReasonCode) == "" {
				out.ReasonCode = "unavailable"
			}
			return out, ErrTrackUnavailable
		}
		return out, nil
	default:
		if resp.StatusCode >= http.StatusInternalServerError {
			return LookupResponse{}, ErrDependencyUnavailable
		}
		return LookupResponse{}, fmt.Errorf("unexpected catalog status: %d", resp.StatusCode)
	}
}
