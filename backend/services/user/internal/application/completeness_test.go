package application

import (
	"testing"

	"jobconnect/user/internal/domain"
)

func baseProfile(role string) domain.Profile {
	return domain.Profile{
		Role:         role,
		DisplayName:  "Jane Doe",
		ContactEmail: "jane@example.com",
		AvatarURL:    "/profiles/u/avatar",
		Bio:          "bio",
		TaxID:        "TIN-123",
	}
}

func completeClient(status string) *domain.ClientProfile {
	return &domain.ClientProfile{
		CompanyName:        "Acme",
		BillingAddress:     "123 St",
		TaxID:              "TIN-123",
		VerificationStatus: status,
	}
}

func completeFreelancer(status string) *domain.FreelancerProfile {
	return &domain.FreelancerProfile{
		Headline:           "Backend Engineer",
		Skills:             []string{"go", "grpc"},
		VerificationStatus: status,
		HourlyRate:         100,
		Availability:       domain.AvailabilityAsNeeded,
	}
}

func hasMissing(missing []string, key string) bool {
	for _, m := range missing {
		if m == key {
			return true
		}
	}
	return false
}

func stepByKey(steps []OnboardingStep, key string) (OnboardingStep, bool) {
	for _, s := range steps {
		if s.Key == key {
			return s, true
		}
	}
	return OnboardingStep{}, false
}

func TestOnboardingUsesStableFourStepKeys(t *testing.T) {
	profile := baseProfile(domain.RoleFreelancer)
	freelancer := completeFreelancer(domain.VerificationStatusSubmitted)

	steps := computeOnboardingSteps(profile, nil, freelancer, readinessSignals{HasWorkPreferences: true})
	if len(steps) != 4 {
		t.Fatalf("expected 4 onboarding steps, got %d", len(steps))
	}

	keys := []string{
		onboardingStepProfile,
		onboardingStepAvatar,
		onboardingStepPreferences,
		onboardingStepKYC,
	}
	for _, key := range keys {
		if _, ok := stepByKey(steps, key); !ok {
			t.Fatalf("missing onboarding step key %q in %v", key, steps)
		}
	}
}

func TestCompletenessClientPendingNeedsKYC(t *testing.T) {
	profile := baseProfile(domain.RoleClient)
	client := completeClient(domain.VerificationStatusPending)

	percent, missing := computeCompleteness(profile, client, nil, readinessSignals{HasHiringPreferences: true})
	if percent != 75 {
		t.Fatalf("expected 75 percent for client with pending verification, got %d", percent)
	}
	if !hasMissing(missing, readinessMissingKYC) {
		t.Fatalf("expected missing %q, got %v", readinessMissingKYC, missing)
	}
}

func TestCompletenessFreelancerSubmittedIsComplete(t *testing.T) {
	profile := baseProfile(domain.RoleFreelancer)
	freelancer := completeFreelancer(domain.VerificationStatusSubmitted)

	percent, missing := computeCompleteness(profile, nil, freelancer, readinessSignals{HasWorkPreferences: true})
	if percent != 100 {
		t.Fatalf("expected 100 percent completeness, got %d", percent)
	}
	if len(missing) != 0 {
		t.Fatalf("expected no missing fields, got %v", missing)
	}
}

func TestCompletenessAdminDoesNotRequireRoleSpecificOrKYC(t *testing.T) {
	profile := baseProfile(domain.RoleAdmin)
	profile.AvatarURL = ""

	percent, missing := computeCompleteness(profile, nil, nil, readinessSignals{})
	if percent != 75 {
		t.Fatalf("expected 75 percent for admin missing only avatar, got %d", percent)
	}
	if !hasMissing(missing, readinessMissingAvatar) {
		t.Fatalf("expected avatar requirement missing, got %v", missing)
	}
	if hasMissing(missing, readinessMissingRoleProfile) {
		t.Fatalf("did not expect role step missing for admin, got %v", missing)
	}
	if hasMissing(missing, readinessMissingKYC) {
		t.Fatalf("did not expect kyc step missing for admin, got %v", missing)
	}
}

func TestCompletenessUsesRequirementStyleMissingKeys(t *testing.T) {
	profile := baseProfile(domain.RoleFreelancer)
	profile.AvatarURL = ""
	freelancer := &domain.FreelancerProfile{}

	_, missing := computeCompleteness(profile, nil, freelancer, readinessSignals{})
	for _, key := range []string{readinessMissingAvatar, readinessMissingCoreProfile, readinessMissingKYC} {
		if !hasMissing(missing, key) {
			t.Fatalf("expected missing %q, got %v", key, missing)
		}
	}
	if !hasMissing(missing, onboardingStepPreferences) {
		t.Fatalf("expected preferences step key in missing list, got %v", missing)
	}
}

func TestReadinessFreelancerIncludesPortfolioPreferencesAndCV(t *testing.T) {
	profile := baseProfile(domain.RoleFreelancer)
	freelancer := completeFreelancer(domain.VerificationStatusSubmitted)

	percent, missing, recs := computeReadiness(profile, nil, freelancer, readinessSignals{})
	if percent != 57 {
		t.Fatalf("expected 57 percent readiness, got %d", percent)
	}
	if !hasMissing(missing, readinessMissingPortfolio) {
		t.Fatalf("expected missing portfolio, got %v", missing)
	}
	if !hasMissing(missing, readinessMissingWorkPreferences) {
		t.Fatalf("expected missing work preferences, got %v", missing)
	}
	if !hasMissing(missing, readinessMissingCV) {
		t.Fatalf("expected missing cv, got %v", missing)
	}
	if len(recs) != len(missing) {
		t.Fatalf("expected recommendations to align with missing list, missing=%v recs=%v", missing, recs)
	}
}

func TestReadinessClientIncludesHiringPreferences(t *testing.T) {
	profile := baseProfile(domain.RoleClient)
	client := completeClient(domain.VerificationStatusSubmitted)

	percent, missing, _ := computeReadiness(profile, client, nil, readinessSignals{})
	if percent != 80 {
		t.Fatalf("expected 80 percent readiness, got %d", percent)
	}
	if !hasMissing(missing, readinessMissingHiringPreferences) {
		t.Fatalf("expected missing hiring preferences, got %v", missing)
	}
}
