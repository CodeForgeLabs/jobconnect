package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Valid roles (must match DB check constraint).
const (
	RoleClient     = "client"
	RoleFreelancer = "freelancer"
	RoleAdmin      = "admin"
)

var validRoles = map[string]bool{
	RoleClient: true, RoleFreelancer: true, RoleAdmin: true,
}

// User is the auth-domain user entity.
type User struct {
	ID              uuid.UUID
	Email           string
	Role            string
	FirstName       string
	LastName        string
	DisplayName     string
	EmailVerifiedAt *time.Time // nil until verified
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ValidateRole returns an error if role is not one of client, freelancer, admin.
func ValidateRole(role string) error {
	if role == "" {
		return fmt.Errorf("role is required")
	}
	if !validRoles[role] {
		return fmt.Errorf("invalid role: %s", role)
	}
	return nil
}

// NormalizeEmail lowercases and trims email.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail returns an error if email is empty or invalid format.
func ValidateEmail(email string) error {
	email = NormalizeEmail(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidateDisplayName returns an error if display name is empty or too long.
func ValidateDisplayName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("display name is required")
	}
	if len(name) > 255 {
		return fmt.Errorf("display name too long")
	}
	return nil
}

// ValidateFirstName returns an error if first name is empty or too long.
func ValidateFirstName(name string) error {
	return validateNamePart(name, "first name")
}

// ValidateLastName returns an error if last name is empty or too long.
func ValidateLastName(name string) error {
	return validateNamePart(name, "last name")
}

func validateNamePart(name, label string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%s is required", label)
	}
	if len(name) > 255 {
		return fmt.Errorf("%s too long", label)
	}
	return nil
}
