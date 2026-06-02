package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DeleteProfileInput struct {
	UserID     uuid.UUID
	HardDelete bool
}

type DeleteProfileOutput struct {
	Deleted bool
}

type DeleteProfile struct {
	Profiles ProfileRepository
	Clock    Clock
}

func (uc *DeleteProfile) Execute(ctx context.Context, in DeleteProfileInput) (DeleteProfileOutput, error) {
	if in.UserID == uuid.Nil {
		return DeleteProfileOutput{}, fmt.Errorf("user_id is required")
	}
	deletedAt := time.Now().UTC()
	if uc.Clock != nil {
		deletedAt = uc.Clock.Now()
	}
	if err := uc.Profiles.Delete(ctx, in.UserID, in.HardDelete, deletedAt); err != nil {
		return DeleteProfileOutput{}, err
	}
	return DeleteProfileOutput{Deleted: true}, nil
}
