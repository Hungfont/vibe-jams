package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken signals a JWT that cannot be verified or decoded.
var ErrInvalidToken = errors.New("invalid token")

// VerifierKey is a named symmetric key used for HS256 verification.
type VerifierKey struct {
	KeyID  string
	Secret string
}

// TokenVerifier performs read-only HS256 JWT verification and returns
// normalized Claims. It holds no signing capability.
type TokenVerifier struct {
	keys map[string][]byte
}

// NewTokenVerifier builds a verifier from the active key and any rotated
// previous keys. All keys must have non-empty KeyID and Secret.
func NewTokenVerifier(active VerifierKey, previous []VerifierKey) (*TokenVerifier, error) {
	if strings.TrimSpace(active.KeyID) == "" {
		return nil, fmt.Errorf("active key kid is required")
	}
	if strings.TrimSpace(active.Secret) == "" {
		return nil, fmt.Errorf("active key secret is required")
	}

	keys := make(map[string][]byte, len(previous)+1)
	keys[active.KeyID] = []byte(active.Secret)
	for _, k := range previous {
		kid := strings.TrimSpace(k.KeyID)
		if kid == "" {
			return nil, fmt.Errorf("previous key kid is required")
		}
		if strings.TrimSpace(k.Secret) == "" {
			return nil, fmt.Errorf("previous key secret is required for kid=%s", kid)
		}
		if _, exists := keys[kid]; exists {
			return nil, fmt.Errorf("duplicate key id %s", kid)
		}
		keys[kid] = []byte(k.Secret)
	}
	return &TokenVerifier{keys: keys}, nil
}

// VerifyAndExtractClaims parses a raw JWT string, verifies the HS256
// signature against the matching kid, and returns validated Claims.
func (v *TokenVerifier) VerifyAndExtractClaims(rawToken string) (Claims, error) {
	parsed, err := jwt.Parse(rawToken, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("%w: unsupported signing method", ErrInvalidToken)
		}
		kid, _ := token.Header["kid"].(string)
		secret, ok := v.keys[strings.TrimSpace(kid)]
		if !ok {
			return nil, fmt.Errorf("%w: unknown kid", ErrInvalidToken)
		}
		return secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	mapClaims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, fmt.Errorf("%w: unexpected claims type", ErrInvalidToken)
	}
	return extractClaims(mapClaims)
}

// extractClaims converts JWT map claims into the shared Claims struct.
func extractClaims(mc jwt.MapClaims) (Claims, error) {
	userID, _ := mc["userId"].(string)
	plan, _ := mc["plan"].(string)
	sessionState, _ := mc["sessionState"].(string)

	if userID == "" || plan == "" || sessionState == "" {
		return Claims{}, fmt.Errorf("%w: missing required fields", ErrInvalidToken)
	}

	var scope []string
	switch raw := mc["scope"].(type) {
	case []string:
		scope = raw
	case []any:
		for _, item := range raw {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				scope = append(scope, s)
			}
		}
	case string:
		if trimmed := strings.TrimSpace(raw); trimmed != "" {
			scope = strings.Fields(trimmed)
		}
	}

	claims := Claims{
		UserID:       userID,
		Plan:         plan,
		SessionState: sessionState,
		Scope:        scope,
	}
	if err := ValidateClaims(claims); err != nil {
		return Claims{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	return claims, nil
}

// ParsePreviousKeys parses a comma-separated "kid:secret,kid2:secret2" string
// into a slice of VerifierKey. Returns nil for empty input.
func ParsePreviousKeys(raw string) ([]VerifierKey, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	pairs := strings.Split(trimmed, ",")
	keys := make([]VerifierKey, 0, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid previous key pair %q", pair)
		}
		keys = append(keys, VerifierKey{
			KeyID:  strings.TrimSpace(parts[0]),
			Secret: strings.TrimSpace(parts[1]),
		})
	}
	return keys, nil
}
