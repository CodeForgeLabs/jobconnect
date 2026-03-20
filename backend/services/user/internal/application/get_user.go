package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"jobconnect/user/internal/domain"
)

type GetUserInput struct {
	UserID uuid.UUID
}

type GetUserOutput struct {
	Profile domain.Profile
}

type GetUser struct {
	Profiles ProfileRepository
}

func (uc *GetUser) Execute(ctx context.Context, in GetUserInput) (GetUserOutput, error) {
	if in.UserID == uuid.Nil {
		return GetUserOutput{}, fmt.Errorf("user_id is required")
	}
	p, _, _, err := uc.Profiles.GetByUserID(ctx, in.UserID)
	if err != nil {
		return GetUserOutput{}, err
	}
	return GetUserOutput{Profile: p}, nil
}
