package application

import (
	"context"
	"fmt"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type GetOnboardingStatusInput struct {
	UserID uuid.UUID
}

type GetOnboardingStatusOutput struct {
	Percent uint32
	Missing []string
	Steps   []OnboardingStep
}

type OnboardingStep struct {
	Key       string
	Completed bool
}

type GetOnboardingStatus struct {
	Profiles ProfileRepository
}

func (uc *GetOnboardingStatus) Build(profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) GetOnboardingStatusOutput {
	percent, missing := computeCompleteness(profile, client, freelancer)
	steps := computeOnboardingSteps(profile, client, freelancer)
	return GetOnboardingStatusOutput{Percent: percent, Missing: missing, Steps: steps}
}

func (uc *GetOnboardingStatus) Execute(ctx context.Context, in GetOnboardingStatusInput) (GetOnboardingStatusOutput, error) {
	if in.UserID == uuid.Nil {
		return GetOnboardingStatusOutput{}, fmt.Errorf("user_id is required")
	}
	profile, client, freelancer, err := uc.Profiles.GetByUserID(ctx, in.UserID)
	if err != nil {
		return GetOnboardingStatusOutput{}, err
	}
	return uc.Build(profile, client, freelancer), nil
}
