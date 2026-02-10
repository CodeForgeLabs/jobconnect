package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// LogoutEverywhereInput is the input (user ID from access token).
type LogoutEverywhereInput struct {
	UserID uuid.UUID
}

// LogoutEverywhere revokes all refresh sessions for the user.
type LogoutEverywhere struct {
	Sessions SessionRepository
}

// Execute revokes all sessions for the given user.
func (uc *LogoutEverywhere) Execute(ctx context.Context, in LogoutEverywhereInput) error {
	if in.UserID == uuid.Nil {
		return fmt.Errorf("user ID required")
	}
	return uc.Sessions.RevokeByUserID(ctx, in.UserID)
}
