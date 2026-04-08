package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testKID = "test-kid"
const testSecret = "test-secret-at-least-32-chars-long"

func signTestToken(t *testing.T, kid, secret string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	return signed
}

func validMapClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"userId":       "u-1",
		"plan":         "premium",
		"sessionState": "valid",
		"scope":        []any{"jam:read", "jam:control"},
		"sid":          "sess-1",
		"kid":          testKID,
		"exp":          float64(time.Now().Add(time.Hour).Unix()),
		"iat":          float64(time.Now().Unix()),
	}
}

func TestNewTokenVerifier_RequiresActiveKey(t *testing.T) {
	t.Parallel()
	_, err := NewTokenVerifier(VerifierKey{KeyID: "", Secret: testSecret}, nil)
	if err == nil {
		t.Fatal("expected error for empty kid")
	}
	_, err = NewTokenVerifier(VerifierKey{KeyID: testKID, Secret: ""}, nil)
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestNewTokenVerifier_RejectsDuplicateKID(t *testing.T) {
	t.Parallel()
	_, err := NewTokenVerifier(
		VerifierKey{KeyID: testKID, Secret: testSecret},
		[]VerifierKey{{KeyID: testKID, Secret: "other"}},
	)
	if err == nil {
		t.Fatal("expected error for duplicate kid")
	}
}

func TestVerifyAndExtractClaims_ValidToken(t *testing.T) {
	t.Parallel()
	v, err := NewTokenVerifier(VerifierKey{KeyID: testKID, Secret: testSecret}, nil)
	if err != nil {
		t.Fatalf("build verifier: %v", err)
	}

	raw := signTestToken(t, testKID, testSecret, validMapClaims())
	claims, err := v.VerifyAndExtractClaims(raw)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.UserID != "u-1" {
		t.Fatalf("userId mismatch: %q", claims.UserID)
	}
	if claims.Plan != "premium" {
		t.Fatalf("plan mismatch: %q", claims.Plan)
	}
	if claims.SessionState != "valid" {
		t.Fatalf("sessionState mismatch: %q", claims.SessionState)
	}
	if len(claims.Scope) != 2 || claims.Scope[0] != "jam:read" {
		t.Fatalf("scope mismatch: %v", claims.Scope)
	}
}

func TestVerifyAndExtractClaims_ExpiredToken(t *testing.T) {
	t.Parallel()
	v, _ := NewTokenVerifier(VerifierKey{KeyID: testKID, Secret: testSecret}, nil)

	mc := validMapClaims()
	mc["exp"] = float64(time.Now().Add(-time.Hour).Unix())

	raw := signTestToken(t, testKID, testSecret, mc)
	_, err := v.VerifyAndExtractClaims(raw)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestVerifyAndExtractClaims_WrongSecret(t *testing.T) {
	t.Parallel()
	v, _ := NewTokenVerifier(VerifierKey{KeyID: testKID, Secret: testSecret}, nil)

	raw := signTestToken(t, testKID, "wrong-secret-totally-different!", validMapClaims())
	_, err := v.VerifyAndExtractClaims(raw)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestVerifyAndExtractClaims_UnknownKID(t *testing.T) {
	t.Parallel()
	v, _ := NewTokenVerifier(VerifierKey{KeyID: testKID, Secret: testSecret}, nil)

	raw := signTestToken(t, "other-kid", testSecret, validMapClaims())
	_, err := v.VerifyAndExtractClaims(raw)
	if err == nil {
		t.Fatal("expected error for unknown kid")
	}
}

func TestVerifyAndExtractClaims_PreviousKeyRotation(t *testing.T) {
	t.Parallel()
	prevKID := "old-kid"
	prevSecret := "old-secret-at-least-32-chars-lo!"
	v, _ := NewTokenVerifier(
		VerifierKey{KeyID: testKID, Secret: testSecret},
		[]VerifierKey{{KeyID: prevKID, Secret: prevSecret}},
	)

	mc := validMapClaims()
	mc["kid"] = prevKID
	raw := signTestToken(t, prevKID, prevSecret, mc)

	claims, err := v.VerifyAndExtractClaims(raw)
	if err != nil {
		t.Fatalf("verify with previous key: %v", err)
	}
	if claims.UserID != "u-1" {
		t.Fatalf("userId mismatch: %q", claims.UserID)
	}
}

func TestVerifyAndExtractClaims_MissingClaims(t *testing.T) {
	t.Parallel()
	v, _ := NewTokenVerifier(VerifierKey{KeyID: testKID, Secret: testSecret}, nil)

	mc := validMapClaims()
	delete(mc, "userId")

	raw := signTestToken(t, testKID, testSecret, mc)
	_, err := v.VerifyAndExtractClaims(raw)
	if err == nil {
		t.Fatal("expected error for missing userId")
	}
}

func TestVerifyAndExtractClaims_InvalidSessionState(t *testing.T) {
	t.Parallel()
	v, _ := NewTokenVerifier(VerifierKey{KeyID: testKID, Secret: testSecret}, nil)

	mc := validMapClaims()
	mc["sessionState"] = "revoked"

	raw := signTestToken(t, testKID, testSecret, mc)
	_, err := v.VerifyAndExtractClaims(raw)
	if err == nil {
		t.Fatal("expected error for invalid sessionState")
	}
}

func TestParsePreviousKeys(t *testing.T) {
	t.Parallel()

	keys, err := ParsePreviousKeys("kid1:secret1,kid2:secret2")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0].KeyID != "kid1" || keys[0].Secret != "secret1" {
		t.Fatalf("first key mismatch: %+v", keys[0])
	}
}

func TestParsePreviousKeys_Empty(t *testing.T) {
	t.Parallel()
	keys, err := ParsePreviousKeys("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if keys != nil {
		t.Fatalf("expected nil for empty input, got %v", keys)
	}
}

func TestParsePreviousKeys_InvalidFormat(t *testing.T) {
	t.Parallel()
	_, err := ParsePreviousKeys("no-colon-here")
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid previous key pair") {
		t.Fatalf("unexpected error: %v", err)
	}
}
