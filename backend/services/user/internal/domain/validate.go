package domain

import (
	"fmt"
	"strings"
)

const (
	MaxAvatarSizeBytes = 5 * 1024 * 1024
)

func ValidateRole(role string) error {
	switch strings.TrimSpace(role) {
	case RoleClient, RoleFreelancer, RoleAdmin:
		return nil
	default:
		return fmt.Errorf("invalid role")
	}
}

func ValidateName(label, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", label)
	}
	return nil
}

func ValidateOptionalName(label, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", label)
	}
	return nil
}

func ValidateDisplayName(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("display_name cannot be empty")
	}
	return nil
}

func ValidateAvatarContentType(contentType string) error {
	ct := strings.TrimSpace(strings.ToLower(contentType))
	switch ct {
	case "image/jpeg", "image/png", "image/webp":
		return nil
	default:
		return fmt.Errorf("unsupported avatar content_type")
	}
}

func ValidateAvatarSize(size int) error {
	if size <= 0 {
		return fmt.Errorf("avatar content is required")
	}
	if size > MaxAvatarSizeBytes {
		return fmt.Errorf("avatar exceeds max size of 5MB")
	}
	return nil
}

func BuildDisplayName(firstName, lastName string) string {
	first := strings.TrimSpace(firstName)
	last := strings.TrimSpace(lastName)
	return strings.TrimSpace(first + " " + last)
}

func ValidateAccountStatus(status string) error {
	normalized := strings.TrimSpace(strings.ToUpper(status))
	normalized = strings.TrimPrefix(normalized, "ACCOUNT_STATUS_")
	switch normalized {
	case AccountStatusActive, AccountStatusSuspended, AccountStatusDeleted:
		return nil
	default:
		return fmt.Errorf("invalid account status")
	}
}

func ValidateProfileVisibility(visibility string) error {
	normalized := strings.TrimSpace(strings.ToUpper(visibility))
	normalized = strings.TrimPrefix(normalized, "PROFILE_VISIBILITY_")
	switch normalized {
	case ProfileVisibilityPublic, ProfileVisibilityPrivate:
		return nil
	default:
		return fmt.Errorf("invalid profile visibility")
	}
}
