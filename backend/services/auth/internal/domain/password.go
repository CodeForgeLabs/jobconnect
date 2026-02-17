package domain

import (
	"fmt"
	"regexp"
	"unicode"
)

// Password policy: min 8 chars, at least one upper, one lower, one number, one special.
var (
	passwordMinLen = 8
	passwordMaxLen = 128
	passwordUpper  = regexp.MustCompile(`[A-Z]`)
	passwordLower  = regexp.MustCompile(`[a-z]`)
	passwordDigit  = regexp.MustCompile(`[0-9]`)
	// Note: cannot include backtick in a raw string literal; this covers common specials.
	passwordSpecial = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?~]`)
)

// ValidatePasswordStrength returns an error if password does not meet policy.
func ValidatePasswordStrength(password string) error {
	if len(password) < passwordMinLen {
		return fmt.Errorf("password must be at least %d characters", passwordMinLen)
	}
	if len(password) > passwordMaxLen {
		return fmt.Errorf("password must be at most %d characters", passwordMaxLen)
	}
	if !passwordUpper.MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !passwordLower.MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !passwordDigit.MatchString(password) {
		return fmt.Errorf("password must contain at least one number")
	}
	hasSpecial := false
	for _, r := range password {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			hasSpecial = true
			break
		}
	}
	if !hasSpecial && !passwordSpecial.MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}
	return nil
}

// PasswordHasher hashes and verifies passwords (port for infrastructure).
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (bool, error)
}
