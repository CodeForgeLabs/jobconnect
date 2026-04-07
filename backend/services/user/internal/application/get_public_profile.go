package application

import (
	"context"
	"fmt"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type GetPublicProfileInput struct {
	UserID uuid.UUID
}

type GetPublicProfileOutput struct {
	Profile    domain.Profile
	Freelancer *domain.FreelancerProfile
}

type GetPublicProfile struct {
	Profiles ProfileRepository
}

func (uc *GetPublicProfile) Execute(ctx context.Context, in GetPublicProfileInput) (GetPublicProfileOutput, error) {
	if in.UserID == uuid.Nil {
		return GetPublicProfileOutput{}, fmt.Errorf("user_id is required")
	}

	profile, _, freelancer, err := uc.Profiles.GetByUserID(ctx, in.UserID)
	if err != nil {
		return GetPublicProfileOutput{}, err
	}
	if profile.Role != domain.RoleFreelancer {
		return GetPublicProfileOutput{}, fmt.Errorf("not found")
	}
	if freelancer == nil {
		return GetPublicProfileOutput{}, fmt.Errorf("not found")
	}

	return GetPublicProfileOutput{Profile: profile, Freelancer: freelancer}, nil
}
