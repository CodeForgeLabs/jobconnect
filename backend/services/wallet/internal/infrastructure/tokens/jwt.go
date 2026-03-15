package tokens

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTParser struct {
	secret []byte
}

func NewJWTParser(secret []byte) *JWTParser {
	return &JWTParser{secret: secret}
}

type accessClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func (p *JWTParser) ParseAccessToken(tokenString string) (uuid.UUID, string, error) {
	tok, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return p.secret, nil
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
