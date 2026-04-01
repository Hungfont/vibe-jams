package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	sharedauth "video-streaming/backend/shared/auth"
)

var (
	// ErrUnauthorized indicates token/session context is unauthorized.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrDependencyUnavailable indicates auth validation dependency is unavailable.
	ErrDependencyUnavailable = errors.New("auth dependency unavailable")
)

// Validator validates bearer tokens and returns normalized claims.
type Validator interface {
	ValidateBearerToken(ctx context.Context, bearerToken string) (sharedauth.Claims, error)
}

// HTTPValidator validates tokens by calling auth-service.
type HTTPValidator struct {
	baseURL string
	client  *http.Client
}

// NewHTTPValidator builds an auth-service HTTP client.
func NewHTTPValidator(baseURL string, timeout time.Duration) *HTTPValidator {
	return &HTTPValidator{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// ValidateBearerToken calls auth-service and validates returned claim contract.
func (v *HTTPValidator) ValidateBearerToken(ctx context.Context, bearerToken string) (sharedauth.Claims, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		v.baseURL+"/internal/v1/auth/validate",
		bytes.NewReader(nil),
	)
	if err != nil {
		return sharedauth.Claims{}, fmt.Errorf("build auth request: %w", err)
	}
	req.Header.Set("Authorization", bearerToken)

	resp, err := v.client.Do(req)
	if err != nil {
		return sharedauth.Claims{}, fmt.Errorf("%w: %v", ErrDependencyUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return sharedauth.Claims{}, ErrUnauthorized
	}
	if resp.StatusCode >= http.StatusInternalServerError {
		return sharedauth.Claims{}, ErrDependencyUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		return sharedauth.Claims{}, fmt.Errorf("unexpected auth-service status: %d", resp.StatusCode)
	}

	var claims sharedauth.Claims
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return sharedauth.Claims{}, fmt.Errorf("decode claims: %w", err)
	}
	if err := sharedauth.ValidateClaims(claims); err != nil {
		return sharedauth.Claims{}, fmt.Errorf("%w: %v", ErrUnauthorized, err)
	}
	return claims, nil
}
