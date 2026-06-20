package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RevokeSessionInput struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

type RevokeSession struct {
	Sessions SessionRepository
	Clock    Clock
}

func (uc *RevokeSession) Execute(ctx context.Context, in RevokeSessionInput) error {
	ownerID, expiresAt, revoked, err := uc.Sessions.GetByID(ctx, in.SessionID)
	if err != nil {
		return err
	}
	if ownerID == uuid.Nil {
		return fmt.Errorf("session not found")
	}
	if ownerID != in.UserID {
		return fmt.Errorf("forbidden session access")
	}
	if revoked {
		return nil
	}
	if uc.Clock.Now().After(expiresAt) {
		// Expired sessions are already effectively dead; still mark revoked for consistency.
		_ = uc.Sessions.UpdateLastUsed(ctx, in.SessionID, time.Now().UTC())
	}
	return uc.Sessions.RevokeByID(ctx, in.SessionID)
}
