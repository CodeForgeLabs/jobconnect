package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type onboardingProfileRepoStub struct {
	profile    domain.Profile
	client     *domain.ClientProfile
	freelancer *domain.FreelancerProfile
	err        error
}

func (s onboardingProfileRepoStub) Create(ctx context.Context, profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) (int64, error) {
	return 0, nil
}

func (s onboardingProfileRepoStub) GetByUserID(ctx context.Context, userID uuid.UUID) (domain.Profile, *domain.ClientProfile, *domain.FreelancerProfile, error) {
	if s.err != nil {
		return domain.Profile{}, nil, nil, s.err
	}
	return s.profile, s.client, s.freelancer, nil
}

func (s onboardingProfileRepoStub) Update(ctx context.Context, profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) error {
	return nil
}

func (s onboardingProfileRepoStub) Delete(ctx context.Context, userID uuid.UUID, hardDelete bool, deletedAt time.Time) error {
	return nil
}

func (s onboardingProfileRepoStub) SaveAvatar(ctx context.Context, avatar domain.Avatar) error {
	return nil
}

func (s onboardingProfileRepoStub) GetAvatar(ctx context.Context, userID uuid.UUID) (domain.Avatar, error) {
	return domain.Avatar{}, nil
}

func (s onboardingProfileRepoStub) RemoveAvatar(ctx context.Context, userID uuid.UUID) error {
	return nil
}

type onboardingDetailsStub struct {
	portfolioResult ListResult[PortfolioItem]
	portfolioErr    error
	workPrefsResult WorkPreferences
	workPrefsErr    error
	hiringPrefs     HiringPreferences
	hiringPrefsErr  error
	cvResult        CV
	cvErr           error
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

func (s onboardingDetailsStub) GetCV(ctx context.Context, userID uuid.UUID) (CV, error) {
	return s.cvResult, s.cvErr
}

func TestComputeReadinessSignals_FreelancerReadsPortfolioAndPreferences(t *testing.T) {
	uc := GetOnboardingStatus{Details: onboardingDetailsStub{
		portfolioResult: ListResult[PortfolioItem]{Items: []PortfolioItem{{ID: 1}}},
		workPrefsResult: WorkPreferences{PreferredProjectLength: "short", ContractTypes: []string{"fixed"}},
		cvResult:        CV{UserID: uuid.New()},
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
	if !signals.HasCV {
		t.Fatalf("expected HasCV to be true")
	}
}

func TestComputeReadinessSignals_ClientReadsHiringPreferences(t *testing.T) {
	uc := GetOnboardingStatus{Details: onboardingDetailsStub{
		hiringPrefs: HiringPreferences{PreferredLocations: []string{"remote"}},
	}}

	signals := uc.computeReadinessSignals(context.Background(), uuid.New(), domain.RoleClient)
	if !signals.HasHiringPreferences {
		t.Fatalf("expected HasHiringPreferences to be true")
	}
	if signals.HasPortfolio || signals.HasWorkPreferences {
		t.Fatalf("did not expect freelancer-specific readiness signals for client role")
	}
	if signals.HasCV {
		t.Fatalf("did not expect HasCV for client role")
	}
}

func TestComputeReadinessSignals_ErrorsDoNotSetSignals(t *testing.T) {
	uc := GetOnboardingStatus{Details: onboardingDetailsStub{
		portfolioErr:   errors.New("portfolio unavailable"),
		workPrefsErr:   errors.New("work preferences unavailable"),
		hiringPrefsErr: errors.New("hiring preferences unavailable"),
		cvErr:          errors.New("cv unavailable"),
	}}

	freelancerSignals := uc.computeReadinessSignals(context.Background(), uuid.New(), domain.RoleFreelancer)
	if freelancerSignals.HasPortfolio || freelancerSignals.HasWorkPreferences || freelancerSignals.HasHiringPreferences || freelancerSignals.HasCV {
		t.Fatalf("expected no readiness signals when detail lookups fail, got %+v", freelancerSignals)
	}

	clientSignals := uc.computeReadinessSignals(context.Background(), uuid.New(), domain.RoleClient)
	if clientSignals.HasPortfolio || clientSignals.HasWorkPreferences || clientSignals.HasHiringPreferences {
		t.Fatalf("expected no readiness signals when detail lookups fail, got %+v", clientSignals)
	}
}

func TestGetOnboardingStatusExecute_FreelancerReadinessUsesDetails(t *testing.T) {
	profile := baseProfile(domain.RoleFreelancer)
	freelancer := completeFreelancer(domain.VerificationStatusSubmitted)

	uc := GetOnboardingStatus{
		Profiles: onboardingProfileRepoStub{
			profile:    profile,
			freelancer: freelancer,
		},
		Details: onboardingDetailsStub{
			portfolioResult: ListResult[PortfolioItem]{Items: []PortfolioItem{{ID: 1}}},
			workPrefsResult: WorkPreferences{PreferredProjectLength: "short", ContractTypes: []string{"fixed"}},
			cvResult:        CV{UserID: uuid.New()},
		},
	}

	out, err := uc.Execute(context.Background(), GetOnboardingStatusInput{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.ReadinessPercent != 100 {
		t.Fatalf("expected readiness percent 100, got %d", out.ReadinessPercent)
	}
	if len(out.ReadinessMissing) != 0 {
		t.Fatalf("expected no readiness missing fields, got %v", out.ReadinessMissing)
	}
}

func TestGetOnboardingStatusExecute_ClientMissingHiringPreferences(t *testing.T) {
	profile := baseProfile(domain.RoleClient)
	client := completeClient(domain.VerificationStatusSubmitted)

	uc := GetOnboardingStatus{
		Profiles: onboardingProfileRepoStub{
			profile: profile,
			client:  client,
		},
		Details: onboardingDetailsStub{},
	}

	out, err := uc.Execute(context.Background(), GetOnboardingStatusInput{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.ReadinessPercent != 80 {
		t.Fatalf("expected readiness percent 80, got %d", out.ReadinessPercent)
	}
	if !hasMissing(out.ReadinessMissing, readinessMissingHiringPreferences) {
		t.Fatalf("expected missing hiring preferences, got %v", out.ReadinessMissing)
	}
}
