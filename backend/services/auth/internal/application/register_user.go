package application

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"jobconnect/auth/internal/domain"

	"github.com/google/uuid"
)

// RegisterUserInput is the input for RegisterUser use-case.
type RegisterUserInput struct {
	Email       string
	Password    string
	FirstName   string
	LastName    string
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
	UserProfiles   UserProfileService
	Connects       ConnectsClient
	Hasher         domain.PasswordHasher
	Clock          Clock
	EmailSend      EmailSender
	Events         UserRegistrationPublisher
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
	if err := domain.ValidateFirstName(in.FirstName); err != nil {
		return RegisterUserOutput{}, err
	}
	if err := domain.ValidateLastName(in.LastName); err != nil {
		return RegisterUserOutput{}, err
	}
	displayName := buildDisplayName(in.FirstName, in.LastName)
	if err := domain.ValidateDisplayName(displayName); err != nil {
		return RegisterUserOutput{}, err
	}
	if err := domain.ValidateRole(in.Role); err != nil {
		return RegisterUserOutput{}, err
	}
	if in.Role == domain.RoleAdmin {
		return RegisterUserOutput{}, fmt.Errorf("admin role cannot be self-registered")
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
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		DisplayName: displayName,
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

	if uc.Events != nil {
		if err := uc.Events.PublishUserRegistered(ctx, CreateProfileInput{
			UserID:      u.ID,
			Role:        u.Role,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			DisplayName: u.DisplayName,
			AvatarURL:   "",
			Email:       u.Email,
		}); err != nil {
			return RegisterUserOutput{}, err
		}
	} else {
		if uc.UserProfiles != nil {
			if err := uc.UserProfiles.CreateProfile(ctx, CreateProfileInput{
				UserID:      u.ID,
				Role:        u.Role,
				FirstName:   u.FirstName,
				LastName:    u.LastName,
				DisplayName: u.DisplayName,
				AvatarURL:   "",
				Email:       u.Email,
			}); err != nil {
				return RegisterUserOutput{}, err
			}
		}
		if uc.Connects != nil && strings.ToLower(in.Role) == "freelancer" {
			if err := uc.Connects.GrantInitialConnects(ctx, u.ID); err != nil {
				fmt.Printf("failed to grant initial connects for user %s: %v\n", u.ID, err)
			}
		}
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
			log.Printf("register otp email send failed email=%s: %v", email, err)
			otpSent = false
		} else {
			otpSent = true
		}
	}

	return RegisterUserOutput{UserID: u.ID, OTPSent: otpSent}, nil
}

func buildDisplayName(firstName, lastName string) string {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	return strings.TrimSpace(firstName + " " + lastName)
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
