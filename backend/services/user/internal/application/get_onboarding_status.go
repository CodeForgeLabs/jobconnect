package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type GetOnboardingStatusInput struct {
	UserID uuid.UUID
}

type GetOnboardingStatusOutput struct {
	Percent uint32
	Missing []string
}

type GetOnboardingStatus struct {
	Profiles ProfileRepository
}

func (uc *GetOnboardingStatus) Execute(ctx context.Context, in GetOnboardingStatusInput) (GetOnboardingStatusOutput, error) {
	if in.UserID == uuid.Nil {
		return GetOnboardingStatusOutput{}, fmt.Errorf("user_id is required")
	}
	profile, client, freelancer, err := uc.Profiles.GetByUserID(ctx, in.UserID)
	if err != nil {
		return GetOnboardingStatusOutput{}, err
	}
	percent, missing := computeCompleteness(profile, client, freelancer)
	return GetOnboardingStatusOutput{Percent: percent, Missing: missing}, nil
}
