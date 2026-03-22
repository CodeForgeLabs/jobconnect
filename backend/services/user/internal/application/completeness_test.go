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

func hasMissing(missing []string, key string) bool {
	for _, m := range missing {
		if m == key {
			return true
		}
	}
	return false
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
