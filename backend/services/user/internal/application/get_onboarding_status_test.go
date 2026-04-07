package application

import (
	"context"
	"errors"
	"testing"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type onboardingDetailsStub struct {
	portfolioResult ListResult[PortfolioItem]
	portfolioErr    error
	workPrefsResult WorkPreferences
	workPrefsErr    error
	hiringPrefs     HiringPreferences
	hiringPrefsErr  error
}

func (s onboardingDetailsStub) ListMyPortfolioItems(ctx context.Context, userID uuid.UUID, pageSize uint32, pageToken string) (ListResult[PortfolioItem], error) {
	return s.portfolioResult, s.portfolioErr
}

func (s onboardingDetailsStub) GetWorkPreferences(ctx context.Context, userID uuid.UUID) (WorkPreferences, error) {
	return s.workPrefsResult, s.workPrefsErr
}

func (s onboardingDetailsStub) GetHiringPreferences(ctx context.Context, userID uuid.UUID) (HiringPreferences, error) {
	return s.hiringPrefs, s.hiringPrefsErr
}

func TestComputeReadinessSignals_FreelancerReadsPortfolioAndPreferences(t *testing.T) {
	uc := GetOnboardingStatus{Details: onboardingDetailsStub{
		portfolioResult: ListResult[PortfolioItem]{Items: []PortfolioItem{{ID: 1}}},
		workPrefsResult: WorkPreferences{PreferredProjectLength: "short", ContractTypes: []string{"fixed"}},
	}}

	signals := uc.computeReadinessSignals(context.Background(), uuid.New(), domain.RoleFreelancer)
	if !signals.HasPortfolio {
		t.Fatalf("expected HasPortfolio to be true")
	}
	if !signals.HasWorkPreferences {
		t.Fatalf("expected HasWorkPreferences to be true")
	}
	if signals.HasHiringPreferences {
		t.Fatalf("did not expect HasHiringPreferences for freelancer role")
	}
}

func TestComputeReadinessSignals_ClientReadsHiringPreferences(t *testing.T) {
	uc := GetOnboardingStatus{Details: onboardingDetailsStub{
		hiringPrefs: HiringPreferences{PreferredExperienceLevels: []string{"senior"}},
	}}

	signals := uc.computeReadinessSignals(context.Background(), uuid.New(), domain.RoleClient)
	if !signals.HasHiringPreferences {
		t.Fatalf("expected HasHiringPreferences to be true")
	}
	if signals.HasPortfolio || signals.HasWorkPreferences {
		t.Fatalf("did not expect freelancer-specific readiness signals for client role")
	}
}

func TestComputeReadinessSignals_ErrorsDoNotSetSignals(t *testing.T) {
	uc := GetOnboardingStatus{Details: onboardingDetailsStub{
		portfolioErr:   errors.New("portfolio unavailable"),
		workPrefsErr:   errors.New("work preferences unavailable"),
		hiringPrefsErr: errors.New("hiring preferences unavailable"),
	}}

	freelancerSignals := uc.computeReadinessSignals(context.Background(), uuid.New(), domain.RoleFreelancer)
	if freelancerSignals.HasPortfolio || freelancerSignals.HasWorkPreferences || freelancerSignals.HasHiringPreferences {
		t.Fatalf("expected no readiness signals when detail lookups fail, got %+v", freelancerSignals)
	}

	clientSignals := uc.computeReadinessSignals(context.Background(), uuid.New(), domain.RoleClient)
	if clientSignals.HasPortfolio || clientSignals.HasWorkPreferences || clientSignals.HasHiringPreferences {
		t.Fatalf("expected no readiness signals when detail lookups fail, got %+v", clientSignals)
	}
}
