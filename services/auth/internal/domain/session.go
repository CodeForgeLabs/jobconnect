package domain

import (
	"github.com/google/uuid"
)

// Session represents a refresh-token session (stored server-side).
type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	CreatedAt        string
	ExpiresAt        string
	RevokedAt        *string
	LastUsedAt       *string
}
