package auth

import (
	"net/http"
	"testing"
)

func TestExtractClaimsFromHeaders_Valid(t *testing.T) {
	t.Parallel()
	h := http.Header{}
	h.Set(HeaderUserID, "u-1")
	h.Set(HeaderPlan, "premium")
	h.Set(HeaderSessionState, "valid")
	h.Set(HeaderScope, "jam:read,jam:control")

	claims, ok := ExtractClaimsFromHeaders(h)
	if !ok {
		t.Fatal("expected ok=true")
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
	if len(claims.Scope) != 2 || claims.Scope[0] != "jam:read" || claims.Scope[1] != "jam:control" {
		t.Fatalf("scope mismatch: %v", claims.Scope)
	}
}

func TestExtractClaimsFromHeaders_NoScope(t *testing.T) {
	t.Parallel()
	h := http.Header{}
	h.Set(HeaderUserID, "u-2")
	h.Set(HeaderPlan, "free")
	h.Set(HeaderSessionState, "valid")

	claims, ok := ExtractClaimsFromHeaders(h)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if claims.UserID != "u-2" {
		t.Fatalf("userId mismatch: %q", claims.UserID)
	}
	if claims.Scope != nil {
		t.Fatalf("expected nil scope, got %v", claims.Scope)
	}
}

func TestExtractClaimsFromHeaders_MissingRequired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header http.Header
	}{
		{"missing userId", func() http.Header {
			h := http.Header{}
			h.Set(HeaderPlan, "free")
			h.Set(HeaderSessionState, "valid")
			return h
		}()},
		{"missing plan", func() http.Header {
			h := http.Header{}
			h.Set(HeaderUserID, "u-1")
			h.Set(HeaderSessionState, "valid")
			return h
		}()},
		{"missing sessionState", func() http.Header {
			h := http.Header{}
			h.Set(HeaderUserID, "u-1")
			h.Set(HeaderPlan, "free")
			return h
		}()},
		{"all empty", http.Header{}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, ok := ExtractClaimsFromHeaders(tt.header)
			if ok {
				t.Fatal("expected ok=false for incomplete headers")
			}
		})
	}
}

func TestExtractClaimsFromHeaders_InvalidSessionState(t *testing.T) {
	t.Parallel()
	h := http.Header{}
	h.Set(HeaderUserID, "u-1")
	h.Set(HeaderPlan, "free")
	h.Set(HeaderSessionState, "revoked")

	_, ok := ExtractClaimsFromHeaders(h)
	if ok {
		t.Fatal("expected ok=false for invalid sessionState")
	}
}
