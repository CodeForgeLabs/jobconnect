package oauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type StateClaims struct {
	Provider string `json:"provider"`
	Role     string `json:"role,omitempty"`
	Nonce    string `json:"nonce"`
	Exp      int64  `json:"exp"`
}

func IssueState(secret []byte, provider, role string, ttl time.Duration) (string, error) {
	nonce, err := randomString(24)
	if err != nil {
		return "", err
	}
	claims := StateClaims{
		Provider: provider,
		Role:     role,
		Nonce:    nonce,
		Exp:      time.Now().UTC().Add(ttl).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	sig := sign(secret, payloadB64)
	return payloadB64 + "." + sig, nil
}

func ParseState(secret []byte, state string, expectedProvider string) (StateClaims, error) {
	parts := strings.Split(state, ".")
	if len(parts) != 2 {
		return StateClaims{}, fmt.Errorf("invalid state")
	}
	payloadB64 := parts[0]
	sig := parts[1]
	if !hmac.Equal([]byte(sign(secret, payloadB64)), []byte(sig)) {
		return StateClaims{}, fmt.Errorf("invalid state signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return StateClaims{}, fmt.Errorf("invalid state payload")
	}
	var claims StateClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return StateClaims{}, fmt.Errorf("invalid state claims")
	}
	if claims.Provider != expectedProvider {
		return StateClaims{}, fmt.Errorf("provider mismatch")
	}
	if time.Now().UTC().Unix() > claims.Exp {
		return StateClaims{}, fmt.Errorf("state expired")
	}
	return claims, nil
}

func sign(secret []byte, payload string) string {
	h := hmac.New(sha256.New, secret)
	_, _ = h.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
