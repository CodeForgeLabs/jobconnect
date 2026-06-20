package application

import (
	"context"
	"fmt"
)

// LogoutInput is the input (refresh token from client).
type LogoutInput struct {
	RefreshToken string
}

// Logout revokes the session tied to the refresh token.
type Logout struct {
	Sessions SessionRepository
	Tokens   TokenIssuer
}

// Execute revokes the session for the given refresh token.
func (uc *Logout) Execute(ctx context.Context, in LogoutInput) error {
	if in.RefreshToken == "" {
		return fmt.Errorf("refresh token required")
	}
	refreshHash, err := uc.Tokens.HashRefreshToken(in.RefreshToken)
	if err != nil {
		return fmt.Errorf("invalid refresh token")
	}
	found, sessionID, _, _, _, err := uc.Sessions.GetByRefreshTokenHash(ctx, refreshHash)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("invalid refresh token")
	}
	return uc.Sessions.RevokeByID(ctx, sessionID)
}
