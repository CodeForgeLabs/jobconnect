package application

import (
	"context"
	"fmt"

	"jobconnect/auth/internal/domain"
)

type ResetPasswordInput struct {
	Email       string
	OTP         string
	NewPassword string
}

type ResetPassword struct {
	Users    UserRepository
	Creds    CredentialRepository
	OTPs     OTPRepository
	Sessions SessionRepository
	Hasher   domain.PasswordHasher
}

func (uc *ResetPassword) Execute(ctx context.Context, in ResetPasswordInput) error {
	if err := domain.ValidateEmail(in.Email); err != nil {
		return err
	}
	if in.OTP == "" {
		return fmt.Errorf("otp is required")
	}
	if err := domain.ValidatePasswordStrength(in.NewPassword); err != nil {
		return err
	}

	email := domain.NormalizeEmail(in.Email)
	user, found, err := uc.Users.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("invalid reset credentials")
	}

	ok, err := uc.OTPs.Consume(ctx, email, domain.OTPPurposeResetPassword, in.OTP, uc.Hasher)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("invalid or expired otp")
	}

	hash, err := uc.Hasher.Hash(in.NewPassword)
	if err != nil {
		return err
	}

	if err := uc.Creds.UpdatePasswordHash(ctx, user.ID, hash); err != nil {
		return err
	}

	if err := uc.Sessions.RevokeByUserID(ctx, user.ID); err != nil {
		return err
	}

	return nil
}
