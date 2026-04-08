package auth

import (
	"net/http"
	"strings"
)

// Standard header names injected by the API gateway after token verification.
const (
	HeaderUserID       = "X-Auth-UserId"
	HeaderPlan         = "X-Auth-Plan"
	HeaderSessionState = "X-Auth-SessionState"
	HeaderScope        = "X-Auth-Scope"
)

// ExtractClaimsFromHeaders reads gateway-injected X-Auth-* headers and
// returns validated Claims. Returns (Claims{}, false) when the required
// headers are missing or the resulting claims fail validation.
func ExtractClaimsFromHeaders(h http.Header) (Claims, bool) {
	userID := strings.TrimSpace(h.Get(HeaderUserID))
	plan := strings.TrimSpace(h.Get(HeaderPlan))
	sessionState := strings.TrimSpace(h.Get(HeaderSessionState))

	if userID == "" || plan == "" || sessionState == "" {
		return Claims{}, false
	}

	var scope []string
	if raw := strings.TrimSpace(h.Get(HeaderScope)); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				scope = append(scope, trimmed)
			}
		}
	}

	claims := Claims{
		UserID:       userID,
		Plan:         plan,
		SessionState: sessionState,
		Scope:        scope,
	}
	if err := ValidateClaims(claims); err != nil {
		return Claims{}, false
	}
	return claims, true
}
