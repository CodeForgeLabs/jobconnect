package application

import (
	"context"
	"time"

	"jobconnect/auth/internal/domain"

	"github.com/google/uuid"
)

// UserRepository persists users.
type UserRepository interface {
	Create(ctx context.Context, u domain.User) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, bool, error)
	SetEmailVerified(ctx context.Context, userID uuid.UUID, at time.Time) error
}

// CreateProfileInput is the input for creating a user profile via user service.
type CreateProfileInput struct {
	UserID      uuid.UUID
	Role        string
	FirstName   string
	LastName    string
	DisplayName string
	AvatarURL   string
}

// UserProfileService creates profiles in the user service.
type UserProfileService interface {
	CreateProfile(ctx context.Context, in CreateProfileInput) error
}

// CredentialRepository stores password hashes per user.
type CredentialRepository interface {
	Create(ctx context.Context, userID uuid.UUID, passwordHash string) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (passwordHash string, found bool, err error)
}

// OTPRepository stores hashed OTPs for email verification / password reset.
type OTPRepository interface {
	Create(ctx context.Context, email, purpose, otpHash string, expiresAt time.Time) error
	Consume(ctx context.Context, email, purpose, otpPlain string, hasher domain.PasswordHasher) (bool, error)
	IncrementAttempts(ctx context.Context, email, purpose string) error
}

// SessionRepository stores refresh sessions.
type SessionRepository interface {
	Create(ctx context.Context, userID uuid.UUID, refreshTokenHash string, expiresAt time.Time) (sessionID uuid.UUID, err error)
	GetByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (found bool, sessionID uuid.UUID, userID uuid.UUID, expiresAt time.Time, revoked bool, err error)
	GetByID(ctx context.Context, sessionID uuid.UUID) (userID uuid.UUID, expiresAt time.Time, revoked bool, err error)
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
	RevokeByID(ctx context.Context, sessionID uuid.UUID) error
	UpdateLastUsed(ctx context.Context, sessionID uuid.UUID, at time.Time) error
}

// Clock provides current time (testable).
type Clock interface {
	Now() time.Time
}

// TokenIssuer issues access (JWT) and refresh (opaque) tokens.
type TokenIssuer interface {
	IssueAccessToken(userID uuid.UUID, role string, expiresIn time.Duration) (string, error)
	IssueRefreshToken() (token string, hash string, err error)
	HashRefreshToken(token string) (hash string, err error)
	ParseAccessToken(token string) (userID uuid.UUID, role string, err error)
}

// TOSRepository records terms/privacy acceptance.
type TOSRepository interface {
	Create(ctx context.Context, userID uuid.UUID, termsVersion, privacyVersion string) error
}

// EmailSender sends emails (e.g. OTP). No-op stub is fine for now.
type EmailSender interface {
	SendVerifyEmailOTP(ctx context.Context, email, otp string) error
}
