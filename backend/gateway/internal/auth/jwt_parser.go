package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type Caller struct {
	UserID  uuid.UUID
	Role    string
	TokenID string
}

type JWTParser struct {
	secret []byte
}

func NewJWTParser(secret []byte) *JWTParser {
	return &JWTParser{secret: secret}
}

func (p *JWTParser) ParseAccessToken(tokenString string) (Caller, error) {
	tok, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return p.secret, nil
	})
	if err != nil {
		return Caller{}, err
	}
	claims, ok := tok.Claims.(*AccessClaims)
	if !ok || !tok.Valid {
		return Caller{}, fmt.Errorf("invalid token")
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return Caller{}, err
	}
	return Caller{UserID: id, Role: claims.Role, TokenID: claims.ID}, nil
}
