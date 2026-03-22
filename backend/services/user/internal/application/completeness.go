package application

import (
	"strings"

	"jobconnect/user/internal/domain"
)

func verificationCountsComplete(status string) bool {
	normalized := strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(status), "VERIFICATION_STATUS_"))
	return normalized == domain.VerificationStatusVerified || normalized == domain.VerificationStatusPending
}

func computeCompleteness(profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) (uint32, []string) {
	required := map[string]string{
		"display_name":  profile.DisplayName,
		"language":      profile.Language,
		"contact_email": profile.ContactEmail,
	}

	switch profile.Role {
	case domain.RoleClient:
		if client != nil {
			required["company_name"] = client.CompanyName
			required["billing_address"] = client.BillingAddress
			required["tax_id"] = client.TaxID
			if verificationCountsComplete(client.VerificationStatus) {
				required["verification_status"] = "filled"
			} else {
				required["verification_status"] = ""
			}
		} else {
			required["company_name"] = ""
			required["billing_address"] = ""
			required["tax_id"] = ""
			required["verification_status"] = ""
		}
	case domain.RoleFreelancer:
		required["bio"] = profile.Bio
		if freelancer != nil {
			required["headline"] = freelancer.Headline
			if len(freelancer.Skills) == 0 {
				required["skills"] = ""
			} else {
				required["skills"] = "filled"
			}
			if verificationCountsComplete(freelancer.VerificationStatus) {
				required["verification_status"] = "filled"
			} else {
				required["verification_status"] = ""
			}
		} else {
			required["headline"] = ""
			required["skills"] = ""
			required["verification_status"] = ""
		}
	case domain.RoleAdmin:
		// Keep admin onboarding intentionally minimal.
	}

	if strings.TrimSpace(profile.AvatarURL) == "" {
		required["avatar"] = ""
	} else {
		required["avatar"] = "filled"
	}

	total := len(required)
	if total == 0 {
		return 100, nil
	}

	missing := make([]string, 0, total)
	filled := 0
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, field)
			continue
		}
		filled++
	}

	return uint32((filled * 100) / total), missing
}

func computeOnboardingSteps(profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) []OnboardingStep {
	steps := []OnboardingStep{
		{Key: "profile_completed", Completed: strings.TrimSpace(profile.DisplayName) != "" && strings.TrimSpace(profile.Language) != "" && strings.TrimSpace(profile.ContactEmail) != ""},
		{Key: "avatar_uploaded", Completed: strings.TrimSpace(profile.AvatarURL) != ""},
	}

	if profile.Role == domain.RoleFreelancer {
		hasSkills := freelancer != nil && len(freelancer.Skills) > 0
		steps = append(steps, OnboardingStep{Key: "skills_added", Completed: hasSkills})
	} else {
		steps = append(steps, OnboardingStep{Key: "company_details_added", Completed: client != nil && strings.TrimSpace(client.CompanyName) != ""})
	}

	return steps
}
