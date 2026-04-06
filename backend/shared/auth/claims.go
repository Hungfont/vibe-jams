package auth

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrMissingUserID indicates the claim payload has no user identity.
	ErrMissingUserID = errors.New("missing userId")
	// ErrMissingPlan indicates the claim payload has no plan for entitlement checks.
	ErrMissingPlan = errors.New("missing plan")
	// ErrInvalidSessionState indicates sessionState is missing or unsupported.
	ErrInvalidSessionState = errors.New("invalid sessionState")
)

const (
	// SessionStateValid marks a claim backed by an active session.
	SessionStateValid = "valid"
	// SessionStateInvalid marks a claim backed by an inactive/revoked session.
	SessionStateInvalid = "invalid"
)

// Claims contains the normalized auth contract used by backend services.
type Claims struct {
	UserID       string   `json:"userId"`
	Plan         string   `json:"plan"`
	SessionState string   `json:"sessionState"`
	Scope        []string `json:"scope,omitempty"`
}

// ValidateClaims ensures required contract fields exist and are valid.
func ValidateClaims(claims Claims) error {
	if strings.TrimSpace(claims.UserID) == "" {
		return ErrMissingUserID
	}
	if strings.TrimSpace(claims.Plan) == "" {
		return ErrMissingPlan
	}

	switch strings.ToLower(strings.TrimSpace(claims.SessionState)) {
	case SessionStateValid, SessionStateInvalid:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidSessionState, claims.SessionState)
	}
}

// IsPremiumPlan maps a plan value to deterministic premium entitlement.
func IsPremiumPlan(plan string) bool {
	switch strings.ToLower(strings.TrimSpace(plan)) {
	case "premium", "premium_plus", "pro":
		return true
	default:
		return false
	}
}
