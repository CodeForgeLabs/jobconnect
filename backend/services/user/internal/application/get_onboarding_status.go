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
	Percent                 uint32
	Missing                 []string
	Steps                   []OnboardingStep
	ReadinessPercent        uint32
	ReadinessMissing        []string
	ReadinessRecommendations []string
}

type OnboardingStep struct {
	Key       string
	Completed bool
}

type GetOnboardingStatus struct {
	Profiles ProfileRepository
	Details  onboardingDetailsRepository
}

type onboardingDetailsRepository interface {
	ListMyPortfolioItems(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (ListResult[PortfolioItem], error)
	GetWorkPreferences(ctx context.Context, userID uuid.UUID) (WorkPreferences, error)
	GetHiringPreferences(ctx context.Context, userID uuid.UUID) (HiringPreferences, error)
}

func (uc *GetOnboardingStatus) Build(profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) GetOnboardingStatusOutput {
	percent, missing := computeCompleteness(profile, client, freelancer)
	steps := computeOnboardingSteps(profile, client, freelancer)
	readinessPercent, readinessMissing, readinessRecommendations := computeReadiness(profile, client, freelancer, readinessSignals{})
	return GetOnboardingStatusOutput{
		Percent:                  percent,
		Missing:                  missing,
		Steps:                    steps,
		ReadinessPercent:         readinessPercent,
		ReadinessMissing:         readinessMissing,
		ReadinessRecommendations: readinessRecommendations,
	}
}

func (uc *GetOnboardingStatus) Execute(ctx context.Context, in GetOnboardingStatusInput) (GetOnboardingStatusOutput, error) {
	if in.UserID == uuid.Nil {
		return GetOnboardingStatusOutput{}, fmt.Errorf("user_id is required")
	}
	profile, client, freelancer, err := uc.Profiles.GetByUserID(ctx, in.UserID)
	if err != nil {
		return GetOnboardingStatusOutput{}, err
	}

	signals := readinessSignals{}
	if uc.Details != nil {
		signals = uc.computeReadinessSignals(ctx, in.UserID, profile.Role)
	}

	percent, missing := computeCompleteness(profile, client, freelancer)
	steps := computeOnboardingSteps(profile, client, freelancer)
	readinessPercent, readinessMissing, readinessRecommendations := computeReadiness(profile, client, freelancer, signals)

	return GetOnboardingStatusOutput{
		Percent:                  percent,
		Missing:                  missing,
		Steps:                    steps,
		ReadinessPercent:         readinessPercent,
		ReadinessMissing:         readinessMissing,
		ReadinessRecommendations: readinessRecommendations,
	}, nil
}

func (uc *GetOnboardingStatus) computeReadinessSignals(ctx context.Context, userID uuid.UUID, role string) readinessSignals {
	signals := readinessSignals{}

	switch role {
	case domain.RoleFreelancer:
		items, err := uc.Details.ListMyPortfolioItems(ctx, userID, 1, "")
		if err == nil && len(items.Items) > 0 {
			signals.HasPortfolio = true
		}

		prefs, err := uc.Details.GetWorkPreferences(ctx, userID)
		if err == nil && hasWorkPreferencesSet(prefs) {
			signals.HasWorkPreferences = true
		}
	case domain.RoleClient:
		prefs, err := uc.Details.GetHiringPreferences(ctx, userID)
		if err == nil && hasHiringPreferencesSet(prefs) {
			signals.HasHiringPreferences = true
		}
	}

	return signals
}
