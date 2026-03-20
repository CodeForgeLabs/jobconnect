package application

import (
	"context"
	"fmt"
	"time"

	"jobconnect/auth/internal/domain"
)

// LoginInput is the input for Login use-case.
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput is the output (tokens).
type LoginOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresInSec int64
}

// Login verifies credentials and issues access + refresh tokens.
type Login struct {
	Users      UserRepository
	Creds      CredentialRepository
	Sessions   SessionRepository
	Hasher     domain.PasswordHasher // Ensure domain.PasswordHasher is properly defined or imported
	Tokens     TokenIssuer
	Clock      Clock
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// Execute runs the Login use-case.
func (uc *Login) Execute(ctx context.Context, in LoginInput) (LoginOutput, error) {
	if err := domain.ValidateEmail(in.Email); err != nil {
		return LoginOutput{}, err
	}

	email := domain.NormalizeEmail(in.Email)
	user, found, err := uc.Users.GetByEmail(ctx, email)
	if err != nil {
		return LoginOutput{}, err
	}
	if !found {
		return LoginOutput{}, fmt.Errorf("invalid email or password")
	}

	hash, found, err := uc.Creds.GetByUserID(ctx, user.ID)
	if err != nil || !found {
		return LoginOutput{}, fmt.Errorf("invalid email or password")
	}
	ok, err := uc.Hasher.Verify(in.Password, hash)
	if err != nil || !ok {
		return LoginOutput{}, fmt.Errorf("invalid email or password")
	}

	refreshToken, refreshHash, err := uc.Tokens.IssueRefreshToken()
	if err != nil {
		return LoginOutput{}, err
	}
	now := uc.Clock.Now()
	expiresAt := now.Add(uc.RefreshTTL)
	_, err = uc.Sessions.Create(ctx, user.ID, refreshHash, expiresAt)
	if err != nil {
		return LoginOutput{}, err
	}

	accessToken, err := uc.Tokens.IssueAccessToken(user.ID, user.Role, uc.AccessTTL)
	if err != nil {
		return LoginOutput{}, err
	}

	return LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresInSec: int64(uc.AccessTTL.Seconds()),
	}, nil
}
