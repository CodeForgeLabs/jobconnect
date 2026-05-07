package application

import (
	"fmt"
	"strings"

	"jobconnect/user/internal/domain"
)

const (
	onboardingStepProfile     = "profile"
	onboardingStepAvatar      = "avatar"
	onboardingStepPreferences = "preferences"
	onboardingStepKYC         = "kyc"

	readinessMissingCoreProfile       = "core_profile"
	readinessMissingAvatar            = "avatar"
	readinessMissingRoleProfile       = "role_profile"
	readinessMissingKYC               = "kyc"
	readinessMissingPortfolio         = "portfolio"
	readinessMissingWorkPreferences   = "work_preferences"
	readinessMissingHiringPreferences = "hiring_preferences"
	readinessMissingCV                = "cv"
)

type readinessSignals struct {
	HasPortfolio         bool
	HasWorkPreferences   bool
	HasHiringPreferences bool
	HasCV                bool
}

func verificationCountsComplete(status string) bool {
	normalized := strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(status), "VERIFICATION_STATUS_"))
	return normalized == domain.VerificationStatusVerified || normalized == domain.VerificationStatusSubmitted || normalized == "PENDING_REVIEW"
}

func computeCompleteness(profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile, signals readinessSignals) (uint32, []string) {
	steps := computeOnboardingSteps(profile, client, freelancer, signals)
	if len(steps) == 0 {
		return 100, nil
	}

	missing := make([]string, 0, len(steps))
	completed := 0
	for _, step := range steps {
		if step.Completed {
			completed++
			continue
		}
		if requirementKey := completenessRequirementKey(step.Key); requirementKey != "" {
			missing = append(missing, requirementKey)
			continue
		}
		missing = append(missing, step.Key)
	}

	return uint32((completed * 100) / len(steps)), missing
}

func computeOnboardingSteps(profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile, signals readinessSignals) []OnboardingStep {
	preferencesComplete := true
	switch profile.Role {
	case domain.RoleFreelancer:
		preferencesComplete = signals.HasWorkPreferences
	case domain.RoleClient:
		preferencesComplete = signals.HasHiringPreferences
	}

	return []OnboardingStep{
		{Key: onboardingStepProfile, Completed: hasCoreProfile(profile) && hasRoleProfile(profile, client, freelancer)},
		{Key: onboardingStepAvatar, Completed: hasAvatar(profile)},
		{Key: onboardingStepPreferences, Completed: preferencesComplete},
		{Key: onboardingStepKYC, Completed: hasKYC(profile, client, freelancer)},
	}
}

func computeReadiness(profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile, signals readinessSignals) (uint32, []string, []string) {
	required := make([]string, 0, 8)
	if !hasCoreProfile(profile) {
		required = append(required, readinessMissingCoreProfile)
	}
	if !hasAvatar(profile) {
		required = append(required, readinessMissingAvatar)
	}
	if !hasRoleProfile(profile, client, freelancer) {
		required = append(required, readinessMissingRoleProfile)
	}
	if !hasKYC(profile, client, freelancer) {
		required = append(required, readinessMissingKYC)
	}

	switch profile.Role {
	case domain.RoleFreelancer:
		if !signals.HasPortfolio {
			required = append(required, readinessMissingPortfolio)
		}
		if !signals.HasWorkPreferences {
			required = append(required, readinessMissingWorkPreferences)
		}
		if !signals.HasCV {
			required = append(required, readinessMissingCV)
		}
	case domain.RoleClient:
		if !signals.HasHiringPreferences {
			required = append(required, readinessMissingHiringPreferences)
		}
	}

	weights := 4
	switch profile.Role {
	case domain.RoleFreelancer:
		weights = 7
	case domain.RoleClient:
		weights = 5
	}

	completed := weights - len(required)
	if completed < 0 {
		completed = 0
	}
	percent := uint32((completed * 100) / weights)

	recommendations := make([]string, 0, len(required))
	for _, key := range required {
		recommendations = append(recommendations, readinessRecommendation(key))
	}

	return percent, required, recommendations
}

func completenessRequirementKey(stepKey string) string {
	switch stepKey {
	case onboardingStepProfile:
		return readinessMissingCoreProfile
	case onboardingStepAvatar:
		return readinessMissingAvatar
	case onboardingStepPreferences:
		return ""
	case onboardingStepKYC:
		return readinessMissingKYC
	default:
		return ""
	}
}

func hasCoreProfile(profile domain.Profile) bool {
	return strings.TrimSpace(profile.DisplayName) != "" && strings.TrimSpace(profile.ContactEmail) != ""
}

func hasAvatar(profile domain.Profile) bool {
	return strings.TrimSpace(profile.AvatarURL) != ""
}

func hasRoleProfile(profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) bool {
	switch profile.Role {
	case domain.RoleFreelancer:
		return freelancer != nil && strings.TrimSpace(freelancer.Headline) != "" && len(freelancer.Skills) > 0
	case domain.RoleClient:
		return client != nil && strings.TrimSpace(client.CompanyName) != ""
	default:
		return true
	}
}

func hasKYC(profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) bool {
	switch profile.Role {
	case domain.RoleFreelancer:
		if strings.TrimSpace(profile.TaxID) == "" {
			return false
		}
		return freelancer != nil && verificationCountsComplete(freelancer.VerificationStatus)
	case domain.RoleClient:
		if client == nil || strings.TrimSpace(client.TaxID) == "" {
			return false
		}
		return verificationCountsComplete(client.VerificationStatus)
	default:
		return true
	}
}

func readinessRecommendation(missingKey string) string {
	switch missingKey {
	case readinessMissingCoreProfile:
		return "Complete display name and contact email."
	case readinessMissingAvatar:
		return "Upload a profile avatar."
	case readinessMissingRoleProfile:
		return "Complete profile details."
	case readinessMissingKYC:
		return "Complete KYC verification and required tax fields."
	case readinessMissingPortfolio:
		return "Add at least one portfolio item."
	case readinessMissingWorkPreferences:
		return "Set work preferences."
	case readinessMissingHiringPreferences:
		return "Set hiring preferences."
	case readinessMissingCV:
		return "Upload your CV."
	default:
		return fmt.Sprintf("Complete missing requirement: %s", missingKey)
	}
}

func hasWorkPreferencesSet(in WorkPreferences) bool {
	return IsProjectLengthPreferenceSet(in.PreferredProjectLength) || in.MinBudgetUSD > 0 || in.MaxBudgetUSD > 0 || len(in.ContractTypes) > 0 || in.WeeklyCapacityHours > 0
}

func hasHiringPreferencesSet(in HiringPreferences) bool {
	if in.MinHourlyRate < 0 || in.MaxHourlyRate < 0 {
		return false
	}
	if in.MaxHourlyRate > 0 && in.MinHourlyRate > in.MaxHourlyRate {
		return false
	}

	hasRateFilter := in.MinHourlyRate > 0 || in.MaxHourlyRate > 0
	for _, location := range in.PreferredLocations {
		if strings.TrimSpace(location) != "" {
			return true
		}
	}

	return hasRateFilter
}
