package application

import (
	"context"
	"fmt"

	"jobconnect/auth/internal/domain"
)

// VerifyEmailOTPInput is the input for VerifyEmailOTP use-case.
type VerifyEmailOTPInput struct {
	Email string
	OTP   string
}

// VerifyEmailOTP marks the user's email as verified if OTP is valid.
type VerifyEmailOTP struct {
	Users  UserRepository
	OTPs   OTPRepository
	Hasher domain.PasswordHasher
	Clock  Clock
}

// Execute runs the VerifyEmailOTP use-case.
func (uc *VerifyEmailOTP) Execute(ctx context.Context, in VerifyEmailOTPInput) (bool, error) {
	if err := domain.ValidateEmail(in.Email); err != nil {
		return false, err
	}
	if len(in.OTP) != domain.OTPLength {
		return false, fmt.Errorf("invalid OTP length")
	}

	email := domain.NormalizeEmail(in.Email)
	ok, err := uc.OTPs.Consume(ctx, email, domain.OTPPurposeVerifyEmail, in.OTP, uc.Hasher)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, fmt.Errorf("invalid or expired OTP")
	}

	user, found, err := uc.Users.GetByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	if !found {
		return false, fmt.Errorf("user not found")
	}

	now := uc.Clock.Now()
	if err := uc.Users.SetEmailVerified(ctx, user.ID, now); err != nil {
		return false, err
	}
	return true, nil
}
