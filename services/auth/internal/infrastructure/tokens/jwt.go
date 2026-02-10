package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTIssuer implements application.TokenIssuer with JWT access tokens and opaque refresh tokens.
type JWTIssuer struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTIssuer returns an issuer using the given secret for signing.
func NewJWTIssuer(secret []byte) *JWTIssuer {
	return &JWTIssuer{secret: secret}
}

type accessClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// IssueAccessToken returns a signed JWT with user ID and role.
func (i *JWTIssuer) IssueAccessToken(userID uuid.UUID, role string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
		UserID: userID.String(),
		Role:   role,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(i.secret)
}

// IssueRefreshToken returns a new opaque token and its hash for storage.
func (i *JWTIssuer) IssueRefreshToken() (token string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	token = base64.URLEncoding.EncodeToString(b)
	hash, err = i.HashRefreshToken(token)
	if err != nil {
		return "", "", err
	}
	return token, hash, nil
}

// HashRefreshToken returns the hash of the refresh token for DB lookup.
func (i *JWTIssuer) HashRefreshToken(token string) (string, error) {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:]), nil
}

// ParseAccessToken returns user ID and role from a JWT.
func (i *JWTIssuer) ParseAccessToken(tokenString string) (uuid.UUID, string, error) {
	tok, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil {
		return uuid.Nil, "", err
	}
	claims, ok := tok.Claims.(*accessClaims)
	if !ok || !tok.Valid {
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, "", err
	}
	return id, claims.Role, nil
}
