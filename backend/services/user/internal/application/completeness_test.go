package application

import (
	"testing"

	"jobconnect/user/internal/domain"
)

func baseProfile(role string) domain.Profile {
	return domain.Profile{
		Role:         role,
		DisplayName:  "Jane Doe",
		Language:     "en",
		ContactEmail: "jane@example.com",
		AvatarURL:    "/profiles/u/avatar",
		Bio:          "bio",
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

func TestCompletenessClientVerifiedCountsComplete(t *testing.T) {
	profile := baseProfile(domain.RoleClient)
	client := completeClient(domain.VerificationStatusVerified)

	percent, missing := computeCompleteness(profile, client, nil)
	if percent != 100 {
		t.Fatalf("expected 100 percent completeness, got %d", percent)
	}
	if hasMissing(missing, "verification_status") {
		t.Fatalf("did not expect verification_status in missing fields: %v", missing)
	}
}

func TestCompletenessClientPendingCountsComplete(t *testing.T) {
	profile := baseProfile(domain.RoleClient)
	client := completeClient(domain.VerificationStatusPending)

	percent, missing := computeCompleteness(profile, client, nil)
	if percent != 100 {
		t.Fatalf("expected 100 percent completeness for pending verification, got %d", percent)
	}
	if hasMissing(missing, "verification_status") {
		t.Fatalf("did not expect verification_status in missing fields: %v", missing)
	}
}

func TestCompletenessClientRejectedIsMissingVerification(t *testing.T) {
	profile := baseProfile(domain.RoleClient)
	client := completeClient(domain.VerificationStatusRejected)

	percent, missing := computeCompleteness(profile, client, nil)
	if !hasMissing(missing, "verification_status") {
		t.Fatalf("expected verification_status missing for rejected status, percent=%d missing=%v", percent, missing)
	}
}

func TestCompletenessClientExpiredIsMissingVerification(t *testing.T) {
	profile := baseProfile(domain.RoleClient)
	client := completeClient(domain.VerificationStatusExpired)

	percent, missing := computeCompleteness(profile, client, nil)
	if !hasMissing(missing, "verification_status") {
		t.Fatalf("expected verification_status missing for expired status, percent=%d missing=%v", percent, missing)
	}
}

func TestCompletenessFreelancerVerifiedCountsComplete(t *testing.T) {
	profile := baseProfile(domain.RoleFreelancer)
	freelancer := completeFreelancer(domain.VerificationStatusVerified)

	percent, missing := computeCompleteness(profile, nil, freelancer)
	if percent != 100 {
		t.Fatalf("expected 100 percent completeness, got %d", percent)
	}
	if hasMissing(missing, "verification_status") {
		t.Fatalf("did not expect verification_status in missing fields: %v", missing)
	}
}

func TestCompletenessFreelancerPendingCountsComplete(t *testing.T) {
	profile := baseProfile(domain.RoleFreelancer)
	freelancer := completeFreelancer(domain.VerificationStatusPending)

	percent, missing := computeCompleteness(profile, nil, freelancer)
	if percent != 100 {
		t.Fatalf("expected 100 percent completeness for pending verification, got %d", percent)
	}
	if hasMissing(missing, "verification_status") {
		t.Fatalf("did not expect verification_status in missing fields: %v", missing)
	}
}

func TestCompletenessFreelancerRejectedIsMissingVerification(t *testing.T) {
	profile := baseProfile(domain.RoleFreelancer)
	freelancer := completeFreelancer(domain.VerificationStatusRejected)

	percent, missing := computeCompleteness(profile, nil, freelancer)
	if !hasMissing(missing, "verification_status") {
		t.Fatalf("expected verification_status missing for rejected status, percent=%d missing=%v", percent, missing)
	}
}

func TestCompletenessFreelancerExpiredIsMissingVerification(t *testing.T) {
	profile := baseProfile(domain.RoleFreelancer)
	freelancer := completeFreelancer(domain.VerificationStatusExpired)

	percent, missing := computeCompleteness(profile, nil, freelancer)
	if !hasMissing(missing, "verification_status") {
		t.Fatalf("expected verification_status missing for expired status, percent=%d missing=%v", percent, missing)
	}
}

func TestCompletenessAdminDoesNotRequireVerification(t *testing.T) {
	profile := baseProfile(domain.RoleAdmin)

	percent, missing := computeCompleteness(profile, nil, nil)
	if percent != 100 {
		t.Fatalf("expected 100 percent completeness for admin baseline, got %d", percent)
	}
	if hasMissing(missing, "verification_status") {
		t.Fatalf("did not expect verification_status in missing fields for admin: %v", missing)
	}
}

func TestOnboardingClientIncludesKYCStep(t *testing.T) {
	profile := baseProfile(domain.RoleClient)
	client := completeClient(domain.VerificationStatusPending)

	steps := computeOnboardingSteps(profile, client, nil)
	step, ok := stepByKey(steps, "kyc_verified")
	if !ok {
		t.Fatalf("expected kyc_verified onboarding step, got steps=%v", steps)
	}
	if !step.Completed {
		t.Fatalf("expected kyc_verified step completed for pending status")
	}
}

func TestOnboardingFreelancerIncludesKYCStep(t *testing.T) {
	profile := baseProfile(domain.RoleFreelancer)
	freelancer := completeFreelancer(domain.VerificationStatusVerified)

	steps := computeOnboardingSteps(profile, nil, freelancer)
	step, ok := stepByKey(steps, "kyc_verified")
	if !ok {
		t.Fatalf("expected kyc_verified onboarding step, got steps=%v", steps)
	}
	if !step.Completed {
		t.Fatalf("expected kyc_verified step completed for verified status")
	}
}
