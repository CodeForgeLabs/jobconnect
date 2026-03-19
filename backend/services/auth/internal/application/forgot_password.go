package application

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"jobconnect/auth/internal/domain"
)

type ForgotPasswordInput struct {
	Email string
}

type ForgotPasswordOutput struct {
	Accepted bool
}

type ForgotPassword struct {
	Users     UserRepository
	OTPs      OTPRepository
	Hasher    domain.PasswordHasher
	Clock     Clock
	EmailSend EmailSender
	OTPTTL    time.Duration
}

func (uc *ForgotPassword) Execute(ctx context.Context, in ForgotPasswordInput) (ForgotPasswordOutput, error) {
	if err := domain.ValidateEmail(in.Email); err != nil {
		return ForgotPasswordOutput{}, err
	}

	email := domain.NormalizeEmail(in.Email)
	_, found, err := uc.Users.GetByEmail(ctx, email)
	if err != nil {
		return ForgotPasswordOutput{}, err
	}
	if !found {
		return ForgotPasswordOutput{Accepted: true}, nil
	}

	otp, err := generateOTPCode(domain.OTPLength)
	if err != nil {
		return ForgotPasswordOutput{}, err
	}
	otpHash, err := uc.Hasher.Hash(otp)
	if err != nil {
		return ForgotPasswordOutput{}, err
	}

	expiresAt := uc.Clock.Now().Add(uc.OTPTTL)
	if err := uc.OTPs.Create(ctx, email, domain.OTPPurposeResetPassword, otpHash, expiresAt); err != nil {
		return ForgotPasswordOutput{}, err
	}

	if uc.EmailSend != nil {
		if err := uc.EmailSend.SendPasswordResetOTP(ctx, email, otp); err != nil {
			log.Printf("forgot-password otp email send failed email=%s: %v", email, err)
		}
	}

	return ForgotPasswordOutput{Accepted: true}, nil
}

func generateOTPCode(length int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("generate otp: %w", err)
		}
		b[i] = digits[n.Int64()]
	}
	return string(b), nil
}
