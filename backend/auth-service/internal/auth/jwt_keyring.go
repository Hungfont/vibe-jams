package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// SigningKey is a JWT symmetric key bound to a kid.
type SigningKey struct {
	KeyID  string
	Secret string
}

// KeyRing signs with the active key and verifies with active + previous keys.
type KeyRing struct {
	active SigningKey
	keys   map[string][]byte
}

func NewKeyRing(active SigningKey, previous []SigningKey) (*KeyRing, error) {
	if strings.TrimSpace(active.KeyID) == "" {
		return nil, fmt.Errorf("active key kid is required")
	}
	if strings.TrimSpace(active.Secret) == "" {
		return nil, fmt.Errorf("active key secret is required")
	}

	keys := make(map[string][]byte, len(previous)+1)
	keys[active.KeyID] = []byte(active.Secret)
	for _, key := range previous {
		kid := strings.TrimSpace(key.KeyID)
		if kid == "" {
			return nil, fmt.Errorf("previous key kid is required")
		}
		if strings.TrimSpace(key.Secret) == "" {
			return nil, fmt.Errorf("previous key secret is required for kid=%s", kid)
		}
		if _, exists := keys[kid]; exists {
			return nil, fmt.Errorf("duplicate key id %s", kid)
		}
		keys[kid] = []byte(key.Secret)
	}

	return &KeyRing{active: active, keys: keys}, nil
}

func ParsePreviousSigningKeys(raw string) ([]SigningKey, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	pairs := strings.Split(trimmed, ",")
	keys := make([]SigningKey, 0, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid previous key pair %q", pair)
		}
		keys = append(keys, SigningKey{KeyID: strings.TrimSpace(parts[0]), Secret: strings.TrimSpace(parts[1])})
	}
	return keys, nil
}

func (k *KeyRing) ActiveKeyID() string {
	return k.active.KeyID
}

func (k *KeyRing) SignAccessToken(claims AccessTokenClaims) (string, error) {
	now := time.Now().Unix()
	if claims.IssuedAt == 0 {
		claims.IssuedAt = now
	}
	if claims.ExpiresAt == 0 {
		return "", fmt.Errorf("exp is required")
	}
	claims.KeyID = k.active.KeyID

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":       claims.UserID,
		"plan":         claims.Plan,
		"sessionState": claims.SessionState,
		"scope":        cloneScope(claims.Scope),
		"sid":          claims.SessionID,
		"kid":          claims.KeyID,
		"exp":          claims.ExpiresAt,
		"iat":          claims.IssuedAt,
	})
	token.Header["kid"] = k.active.KeyID

	signed, err := token.SignedString(k.keys[k.active.KeyID])
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (k *KeyRing) VerifyAccessToken(raw string, now time.Time) (AccessTokenClaims, error) {
	parsed, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("%w: unsupported signing method", ErrInvalidToken)
		}
		kid, _ := token.Header["kid"].(string)
		secret, ok := k.keys[strings.TrimSpace(kid)]
		if !ok {
			return nil, fmt.Errorf("%w: unknown kid", ErrInvalidToken)
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithTimeFunc(func() time.Time { return now }))
	if err != nil {
		return AccessTokenClaims{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	mapClaims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return AccessTokenClaims{}, fmt.Errorf("%w: invalid claims", ErrInvalidToken)
	}
	claims, err := mapClaimsToAccessClaims(mapClaims)
	if err != nil {
		return AccessTokenClaims{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	return claims, nil
}

func mapClaimsToAccessClaims(claims jwt.MapClaims) (AccessTokenClaims, error) {
	userID, _ := claims["userId"].(string)
	plan, _ := claims["plan"].(string)
	sessionState, _ := claims["sessionState"].(string)
	sid, _ := claims["sid"].(string)
	kid, _ := claims["kid"].(string)
	if userID == "" || plan == "" || sessionState == "" || sid == "" || kid == "" {
		return AccessTokenClaims{}, fmt.Errorf("missing required claims")
	}

	exp, err := extractUnixClaim(claims, "exp")
	if err != nil {
		return AccessTokenClaims{}, err
	}
	iat, _ := extractUnixClaim(claims, "iat")

	var scope []string
	switch raw := claims["scope"].(type) {
	case []string:
		scope = cloneScope(raw)
	case []any:
		scope = make([]string, 0, len(raw))
		for _, item := range raw {
			str, ok := item.(string)
			if !ok || strings.TrimSpace(str) == "" {
				continue
			}
			scope = append(scope, str)
		}
	case string:
		if strings.TrimSpace(raw) != "" {
			scope = strings.Fields(raw)
		}
	}

	return AccessTokenClaims{
		UserID:       userID,
		Plan:         plan,
		SessionState: sessionState,
		Scope:        scope,
		SessionID:    sid,
		KeyID:        kid,
		ExpiresAt:    exp,
		IssuedAt:     iat,
	}, nil
}

func extractUnixClaim(claims jwt.MapClaims, key string) (int64, error) {
	raw, ok := claims[key]
	if !ok {
		return 0, fmt.Errorf("missing %s", key)
	}
	switch value := raw.(type) {
	case float64:
		return int64(value), nil
	case int64:
		return value, nil
	case jsonNumber:
		result, err := value.Int64()
		if err != nil {
			return 0, fmt.Errorf("invalid %s", key)
		}
		return result, nil
	default:
		return 0, fmt.Errorf("invalid %s", key)
	}
}

type jsonNumber interface {
	Int64() (int64, error)
}
