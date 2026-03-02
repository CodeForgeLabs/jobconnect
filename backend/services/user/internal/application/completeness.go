package application

import (
	"strings"

	"jobconnect/user/internal/domain"
)

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
		} else {
			required["company_name"] = ""
			required["billing_address"] = ""
			required["tax_id"] = ""
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
		} else {
			required["headline"] = ""
			required["skills"] = ""
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
