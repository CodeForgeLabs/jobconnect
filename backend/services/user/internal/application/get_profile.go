package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"jobconnect/user/internal/domain"
)

type GetProfileInput struct {
	UserID uuid.UUID
}

type GetProfileOutput struct {
	Profile    domain.Profile
	Client     *domain.ClientProfile
	Freelancer *domain.FreelancerProfile
}

type GetProfile struct {
	Profiles ProfileRepository
}

func (uc *GetProfile) Execute(ctx context.Context, in GetProfileInput) (GetProfileOutput, error) {
	if in.UserID == uuid.Nil {
		return GetProfileOutput{}, fmt.Errorf("user_id is required")
	}
	p, client, freelancer, err := uc.Profiles.GetByUserID(ctx, in.UserID)
	if err != nil {
		return GetProfileOutput{}, err
	}
	return GetProfileOutput{Profile: p, Client: client, Freelancer: freelancer}, nil
}
