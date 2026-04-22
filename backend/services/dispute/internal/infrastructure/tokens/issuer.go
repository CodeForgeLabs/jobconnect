package tokens

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTIssuer struct {
	secret []byte
}

func NewJWTIssuer(secret []byte) *JWTIssuer {
	return &JWTIssuer{secret: secret}
}

func (i *JWTIssuer) IssueAccessToken(userID uuid.UUID, role string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		UserID: userID.String(),
		Role:   role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(i.secret)
}
