package challenge

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"
)

type proofClaims struct {
	IP  string `json:"ip"`
	Exp int64  `json:"exp"`
}

func IssueProof(secret []byte, ip string, ttl time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(ttl)
	claims := proofClaims{IP: ip, Exp: expiresAt.Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	sig := sign(secret, payloadB64)
	return payloadB64 + "." + sig, expiresAt, nil
}

func VerifyProof(secret []byte, proof string, ip string, now time.Time) bool {
	parts := split2(proof, '.')
	if len(parts) != 2 {
		return false
	}
	payloadB64 := parts[0]
	sig := parts[1]
	if !hmac.Equal([]byte(sign(secret, payloadB64)), []byte(sig)) {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return false
	}
	var claims proofClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return false
	}
	if claims.IP == "" || claims.IP != ip {
		return false
	}
	if now.UTC().Unix() > claims.Exp {
		return false
	}
	return true
}

func sign(secret []byte, payload string) string {
	h := hmac.New(sha256.New, secret)
	_, _ = h.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func split2(s string, sep byte) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return nil
}
