package tokens

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTIssuer_IssueAndParseAccessToken(t *testing.T) {
	issuer := NewJWTIssuer([]byte("test-secret"))
	userID := uuid.New()
	role := "client"

	tok, err := issuer.IssueAccessToken(userID, role, time.Minute)
	if err != nil {
		t.Fatalf("IssueAccessToken error: %v", err)
	}
	gotID, gotRole, err := issuer.ParseAccessToken(tok)
	if err != nil {
		t.Fatalf("ParseAccessToken error: %v", err)
	}
	if gotID != userID {
		t.Fatalf("userID mismatch: got %v, want %v", gotID, userID)
	}
	if gotRole != role {
		t.Fatalf("role mismatch: got %q, want %q", gotRole, role)
	}
}

func TestJWTIssuer_IssueRefreshToken(t *testing.T) {
	issuer := NewJWTIssuer([]byte("test-secret"))
	token, hash, err := issuer.IssueRefreshToken()
	if err != nil {
		t.Fatalf("IssueRefreshToken error: %v", err)
	}
	if token == "" || hash == "" {
		t.Fatalf("expected non-empty token and hash")
	}
	h2, err := issuer.HashRefreshToken(token)
	if err != nil {
		t.Fatalf("HashRefreshToken error: %v", err)
	}
	if h2 != hash {
		t.Fatalf("hash mismatch")
	}
}
