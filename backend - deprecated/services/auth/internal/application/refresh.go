package application

import (
	"context"
	"fmt"
	"time"
)

// RefreshInput is the input for Refresh use-case.
type RefreshInput struct {
	RefreshToken string
}

// RefreshOutput is the new tokens.
type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresInSec int64
}

// Refresh validates refresh token and issues new access + refresh tokens.
type Refresh struct {
	Users      UserRepository
	Sessions   SessionRepository
	Tokens     TokenIssuer
	Clock      Clock
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// Execute runs the Refresh use-case.
func (uc *Refresh) Execute(ctx context.Context, in RefreshInput) (RefreshOutput, error) {
	if in.RefreshToken == "" {
		return RefreshOutput{}, fmt.Errorf("refresh token required")
	}

	// TokenIssuer should provide a way to hash the refresh token for lookup, or we store it in session.
	// We have IssueRefreshToken() returning (token, hash). So we need HashRefreshToken(token) -> hash for lookup.
	// Add to TokenIssuer: HashRefreshToken(token string) (hash string, err error)
	// For now assume TokenIssuer has HashRefreshToken. I'll add it to the interface.
	refreshHash, err := uc.Tokens.HashRefreshToken(in.RefreshToken)
	if err != nil {
		return RefreshOutput{}, fmt.Errorf("invalid refresh token")
	}

	found, sessionID, userID, expiresAt, revoked, err := uc.Sessions.GetByRefreshTokenHash(ctx, refreshHash)
	if err != nil {
		return RefreshOutput{}, err
	}
	if !found {
		return RefreshOutput{}, fmt.Errorf("invalid refresh token")
	}
	if revoked {
		return RefreshOutput{}, fmt.Errorf("session revoked")
	}
	now := uc.Clock.Now()
	if now.After(expiresAt) {
		return RefreshOutput{}, fmt.Errorf("refresh token expired")
	}

	user, found, err := uc.Users.GetByID(ctx, userID)
	if err != nil || !found {
		return RefreshOutput{}, fmt.Errorf("user not found")
	}

	// Rotate refresh token: create new session, optionally revoke old one (or leave for audit).
	newRefreshToken, newRefreshHash, err := uc.Tokens.IssueRefreshToken()
	if err != nil {
		return RefreshOutput{}, err
	}
	newExpiresAt := now.Add(uc.RefreshTTL)
	_, err = uc.Sessions.Create(ctx, user.ID, newRefreshHash, newExpiresAt)
	if err != nil {
		return RefreshOutput{}, err
	}
	_ = uc.Sessions.RevokeByID(ctx, sessionID)

	accessToken, err := uc.Tokens.IssueAccessToken(user.ID, user.Role, uc.AccessTTL)
	if err != nil {
		return RefreshOutput{}, err
	}

	return RefreshOutput{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresInSec: int64(uc.AccessTTL.Seconds()),
	}, nil
}
