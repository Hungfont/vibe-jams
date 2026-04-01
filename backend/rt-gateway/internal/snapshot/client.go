package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"video-streaming/backend/rt-gateway/internal/model"
)

// Client fetches authoritative jam state snapshots from jam-service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a snapshot client.
func NewClient(baseURL string, timeout time.Duration) *Client {
	trimmed := strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL:    trimmed,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// FetchSessionState returns the current authoritative session+queue snapshot.
func (c *Client) FetchSessionState(ctx context.Context, sessionID string) (model.SessionStateSnapshot, error) {
	if sessionID == "" {
		return model.SessionStateSnapshot{}, fmt.Errorf("sessionID is required")
	}
	if c.baseURL == "" {
		return model.SessionStateSnapshot{}, fmt.Errorf("baseURL is required")
	}

	endpoint := c.baseURL + "/api/v1/jams/" + url.PathEscape(sessionID) + "/state"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return model.SessionStateSnapshot{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.SessionStateSnapshot{}, fmt.Errorf("fetch snapshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.SessionStateSnapshot{}, fmt.Errorf("unexpected snapshot status: %d", resp.StatusCode)
	}

	var snapshot model.SessionStateSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		return model.SessionStateSnapshot{}, fmt.Errorf("decode snapshot: %w", err)
	}

	if snapshot.AggregateVersion <= 0 {
		snapshot.AggregateVersion = snapshot.Session.SessionVersion
		if snapshot.Queue.QueueVersion > snapshot.AggregateVersion {
			snapshot.AggregateVersion = snapshot.Queue.QueueVersion
		}
	}

	return snapshot, nil
}
