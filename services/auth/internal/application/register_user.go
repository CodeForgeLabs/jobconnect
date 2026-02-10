package application

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"jobconnect/auth/internal/domain"

	"github.com/google/uuid"
)

// RegisterUserInput is the input for RegisterUser use-case.
type RegisterUserInput struct {
	Email       string
	Password    string
	DisplayName string
	Role        string
	AcceptTerms bool
}

// RegisterUserOutput is the output of RegisterUser use-case.
type RegisterUserOutput struct {
	UserID  uuid.UUID
	OTPSent bool
}

// RegisterUser creates a user, stores credential, creates OTP and optionally sends email.
type RegisterUser struct {
	Users          UserRepository
	Creds          CredentialRepository
	OTPs           OTPRepository
	TOS            TOSRepository
	Hasher         domain.PasswordHasher
	Clock          Clock
	EmailSend      EmailSender
	OTPTTL         time.Duration
	TOSVersion     string
	PrivacyVersion string
}

// Execute runs the RegisterUser use-case.
func (uc *RegisterUser) Execute(ctx context.Context, in RegisterUserInput) (RegisterUserOutput, error) {
	if !in.AcceptTerms {
		return RegisterUserOutput{}, fmt.Errorf("terms of service and privacy policy must be accepted")
	}
	if err := domain.ValidateEmail(in.Email); err != nil {
		return RegisterUserOutput{}, err
	}
	if err := domain.ValidatePasswordStrength(in.Password); err != nil {
		return RegisterUserOutput{}, err
	}
	if err := domain.ValidateDisplayName(in.DisplayName); err != nil {
		return RegisterUserOutput{}, err
	}
	if err := domain.ValidateRole(in.Role); err != nil {
		return RegisterUserOutput{}, err
	}

	email := domain.NormalizeEmail(in.Email)
	_, found, err := uc.Users.GetByEmail(ctx, email)
	if err != nil {
		return RegisterUserOutput{}, err
	}
	if found {
		return RegisterUserOutput{}, fmt.Errorf("email already registered")
	}

	hash, err := uc.Hasher.Hash(in.Password)
	if err != nil {
		return RegisterUserOutput{}, fmt.Errorf("hashing password: %w", err)
	}

	now := uc.Clock.Now()
	u := domain.User{
		ID:          uuid.New(),
		Email:       email,
		Role:        in.Role,
		DisplayName: in.DisplayName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	u, err = uc.Users.Create(ctx, u)
	if err != nil {
		return RegisterUserOutput{}, err
	}

	if err := uc.Creds.Create(ctx, u.ID, hash); err != nil {
		return RegisterUserOutput{}, err
	}

	if uc.TOS != nil {
		_ = uc.TOS.Create(ctx, u.ID, uc.TOSVersion, uc.PrivacyVersion)
	}

	otp, err := generateOTP(domain.OTPLength)
	if err != nil {
		return RegisterUserOutput{}, err
	}
	otpHash, err := uc.Hasher.Hash(otp)
	if err != nil {
		return RegisterUserOutput{}, err
	}
	expiresAt := now.Add(uc.OTPTTL)
	if err := uc.OTPs.Create(ctx, email, domain.OTPPurposeVerifyEmail, otpHash, expiresAt); err != nil {
		return RegisterUserOutput{}, err
	}

	otpSent := false
	if uc.EmailSend != nil {
		if err := uc.EmailSend.SendVerifyEmailOTP(ctx, email, otp); err != nil {
			// log but don't fail registration
			otpSent = false
		} else {
			otpSent = true
		}
	}

	return RegisterUserOutput{UserID: u.ID, OTPSent: otpSent}, nil
}

func generateOTP(length int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		b[i] = digits[n.Int64()]
	}
	return string(b), nil
}
