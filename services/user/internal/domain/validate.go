package domain

import (
	"fmt"
	"strings"
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

func BuildDisplayName(firstName, lastName string) string {
	first := strings.TrimSpace(firstName)
	last := strings.TrimSpace(lastName)
	return strings.TrimSpace(first + " " + last)
}
