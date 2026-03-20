package application

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"jobconnect/auth/internal/domain"

	"github.com/google/uuid"
)

type OAuthLoginInput struct {
	Provider       string
	ProviderUserID string
	Email          string
	FirstName      string
	LastName       string
	DisplayName    string
	Role           string
}

type OAuthLoginOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresInSec int64
	IsNewUser    bool
}

type OAuthLogin struct {
	Users        UserRepository
	Identities   OAuthIdentityRepository
	Sessions     SessionRepository
	Tokens       TokenIssuer
	Clock        Clock
	UserProfiles UserProfileService
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
}

func (uc *OAuthLogin) Execute(ctx context.Context, in OAuthLoginInput) (OAuthLoginOutput, error) {
	provider := strings.ToLower(strings.TrimSpace(in.Provider))
	if provider == "" {
		return OAuthLoginOutput{}, fmt.Errorf("provider is required")
	}
	providerUserID := strings.TrimSpace(in.ProviderUserID)
	if providerUserID == "" {
		return OAuthLoginOutput{}, fmt.Errorf("provider_user_id is required")
	}
	if err := domain.ValidateEmail(in.Email); err != nil {
		return OAuthLoginOutput{}, err
	}
	email := domain.NormalizeEmail(in.Email)

	identity, foundIdentity, err := uc.Identities.GetByProviderUserID(ctx, provider, providerUserID)
	if err != nil {
		return OAuthLoginOutput{}, err
	}

	var user domain.User
	isNewUser := false
	if foundIdentity {
		var foundUserByID bool
		user, foundUserByID, err = uc.Users.GetByID(ctx, identity.UserID)
		if err != nil {
			return OAuthLoginOutput{}, err
		}
		if !foundUserByID || user.ID == uuid.Nil {
			// Attempt self-heal: resolve by email and relink provider identity.
			var foundByEmail bool
			var userByEmail domain.User
			userByEmail, foundByEmail, err = uc.Users.GetByEmail(ctx, email)
			if err != nil {
				return OAuthLoginOutput{}, err
			}
			if foundByEmail {
				user = userByEmail
			} else {
				role := strings.TrimSpace(in.Role)
				if role == "" {
					role = domain.RoleClient
				}
				if err := domain.ValidateRole(role); err != nil {
					return OAuthLoginOutput{}, err
				}

				firstName := strings.TrimSpace(in.FirstName)
				if firstName == "" {
					firstName = "User"
				}
				lastName := strings.TrimSpace(in.LastName)
				if lastName == "" {
					lastName = "OAuth"
				}
				displayName := strings.TrimSpace(in.DisplayName)
				if displayName == "" {
					displayName = strings.TrimSpace(firstName + " " + lastName)
				}

				now := uc.Clock.Now()
				user, err = uc.Users.Create(ctx, domain.User{
					ID:              uuid.New(),
					Email:           email,
					Role:            role,
					FirstName:       firstName,
					LastName:        lastName,
					DisplayName:     displayName,
					EmailVerifiedAt: &now,
					CreatedAt:       now,
					UpdatedAt:       now,
				})
				if err != nil {
					return OAuthLoginOutput{}, err
				}
				if uc.UserProfiles != nil {
					if err := uc.UserProfiles.CreateProfile(ctx, CreateProfileInput{
						UserID:      user.ID,
						Role:        user.Role,
						FirstName:   user.FirstName,
						LastName:    user.LastName,
						DisplayName: user.DisplayName,
						AvatarURL:   "",
					}); err != nil {
						return OAuthLoginOutput{}, err
					}
				}
				isNewUser = true
			}

			if err := uc.Identities.Create(ctx, OAuthIdentity{
				UserID:         user.ID,
				Provider:       provider,
				ProviderUserID: providerUserID,
				Email:          email,
			}); err != nil {
				return OAuthLoginOutput{}, err
			}
		}
	} else {
		var foundUser bool
		user, foundUser, err = uc.Users.GetByEmail(ctx, email)
		if err != nil {
			return OAuthLoginOutput{}, err
		}
		if !foundUser {
			role := strings.TrimSpace(in.Role)
			if role == "" {
				role = domain.RoleClient
			}
			if err := domain.ValidateRole(role); err != nil {
				return OAuthLoginOutput{}, err
			}

			firstName := strings.TrimSpace(in.FirstName)
			if firstName == "" {
				firstName = "User"
			}
			lastName := strings.TrimSpace(in.LastName)
			if lastName == "" {
				lastName = "OAuth"
			}
			displayName := strings.TrimSpace(in.DisplayName)
			if displayName == "" {
				displayName = strings.TrimSpace(firstName + " " + lastName)
			}

			now := uc.Clock.Now()
			user, err = uc.Users.Create(ctx, domain.User{
				ID:              uuid.New(),
				Email:           email,
				Role:            role,
				FirstName:       firstName,
				LastName:        lastName,
				DisplayName:     displayName,
				EmailVerifiedAt: &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			})
			if err != nil {
				return OAuthLoginOutput{}, err
			}
			if uc.UserProfiles != nil {
				if err := uc.UserProfiles.CreateProfile(ctx, CreateProfileInput{
					UserID:      user.ID,
					Role:        user.Role,
					FirstName:   user.FirstName,
					LastName:    user.LastName,
					DisplayName: user.DisplayName,
					AvatarURL:   "",
				}); err != nil {
					return OAuthLoginOutput{}, err
				}
			}
			isNewUser = true
		}

		if err := uc.Identities.Create(ctx, OAuthIdentity{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: providerUserID,
			Email:          email,
		}); err != nil {
			return OAuthLoginOutput{}, err
		}
	}

	refreshToken, refreshHash, err := uc.Tokens.IssueRefreshToken()
	if err != nil {
		return OAuthLoginOutput{}, err
	}
	now := uc.Clock.Now()
	if _, err := uc.Sessions.Create(ctx, user.ID, refreshHash, now.Add(uc.RefreshTTL)); err != nil {
		log.Printf("oauth session create failed provider=%s provider_user_id=%s user_id=%s email=%s: %v", provider, providerUserID, user.ID.String(), email, err)
		return OAuthLoginOutput{}, err
	}
	accessToken, err := uc.Tokens.IssueAccessToken(user.ID, user.Role, uc.AccessTTL)
	if err != nil {
		return OAuthLoginOutput{}, err
	}

	return OAuthLoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresInSec: int64(uc.AccessTTL.Seconds()),
		IsNewUser:    isNewUser,
	}, nil
}
