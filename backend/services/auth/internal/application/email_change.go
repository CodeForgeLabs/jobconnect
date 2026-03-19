package application

import (
	"context"
	"fmt"
	"time"

	"jobconnect/auth/internal/domain"

	"github.com/google/uuid"
)

type RequestEmailChangeInput struct {
	UserID   uuid.UUID
	NewEmail string
}

type RequestEmailChangeOutput struct {
	OTPSent bool
}

type RequestEmailChange struct {
	Users          UserRepository
	EmailChanges   EmailChangeRequestRepository
	Hasher         domain.PasswordHasher
	Clock          Clock
	EmailSend      EmailSender
	EmailOTPExpiry time.Duration
}

func (uc *RequestEmailChange) Execute(ctx context.Context, in RequestEmailChangeInput) (RequestEmailChangeOutput, error) {
	if in.UserID == uuid.Nil {
		return RequestEmailChangeOutput{}, fmt.Errorf("user_id is required")
	}
	if err := domain.ValidateEmail(in.NewEmail); err != nil {
		return RequestEmailChangeOutput{}, err
	}

	newEmail := domain.NormalizeEmail(in.NewEmail)
	_, found, err := uc.Users.GetByEmail(ctx, newEmail)
	if err != nil {
		return RequestEmailChangeOutput{}, err
	}
	if found {
		return RequestEmailChangeOutput{}, fmt.Errorf("email already registered")
	}

	otp, err := generateOTPCode(domain.OTPLength)
	if err != nil {
		return RequestEmailChangeOutput{}, err
	}
	otpHash, err := uc.Hasher.Hash(otp)
	if err != nil {
		return RequestEmailChangeOutput{}, err
	}
	expiresAt := uc.Clock.Now().Add(uc.EmailOTPExpiry)
	if err := uc.EmailChanges.Upsert(ctx, in.UserID, newEmail, otpHash, expiresAt); err != nil {
		return RequestEmailChangeOutput{}, err
	}

	otpSent := false
	if uc.EmailSend != nil {
		if err := uc.EmailSend.SendEmailChangeOTP(ctx, newEmail, otp); err == nil {
			otpSent = true
		}
	}

	return RequestEmailChangeOutput{OTPSent: otpSent}, nil
}

type ConfirmEmailChangeInput struct {
	UserID uuid.UUID
	OTP    string
}

type ConfirmEmailChange struct {
	Users        UserRepository
	EmailChanges EmailChangeRequestRepository
	Hasher       domain.PasswordHasher
	Clock        Clock
}

func (uc *ConfirmEmailChange) Execute(ctx context.Context, in ConfirmEmailChangeInput) error {
	if in.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if in.OTP == "" {
		return fmt.Errorf("otp is required")
	}

	now := uc.Clock.Now()
	newEmail, ok, err := uc.EmailChanges.Consume(ctx, in.UserID, in.OTP, uc.Hasher, now)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("invalid or expired otp")
	}

	if err := uc.Users.UpdateEmail(ctx, in.UserID, newEmail, now); err != nil {
		return err
	}
	if err := uc.Users.SetEmailVerified(ctx, in.UserID, now); err != nil {
		return err
	}
	if err := uc.EmailChanges.MarkConfirmed(ctx, in.UserID, now); err != nil {
		return err
	}

	return nil
}
