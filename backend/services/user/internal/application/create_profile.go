package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

// CreateProfileInput is the input for CreateProfile use-case.
type CreateProfileInput struct {
	UserID       uuid.UUID
	Role         string
	FirstName    string
	LastName     string
	DisplayName  string
	ContactEmail string
	AvatarURL    string
	Client       *domain.ClientProfile
	Freelancer   *domain.FreelancerProfile
}

// CreateProfileOutput is the output of CreateProfile use-case.
type CreateProfileOutput struct {
	ProfileID int64
}

// CreateProfile creates a base profile and optional role-specific record.
type CreateProfile struct {
	Profiles ProfileRepository
	Clock    Clock
}

// Execute runs the CreateProfile use-case.
func (uc *CreateProfile) Execute(ctx context.Context, in CreateProfileInput) (CreateProfileOutput, error) {
	if in.UserID == uuid.Nil {
		return CreateProfileOutput{}, fmt.Errorf("user_id is required")
	}
	if err := domain.ValidateRole(in.Role); err != nil {
		return CreateProfileOutput{}, err
	}
	if err := domain.ValidateName("first_name", in.FirstName); err != nil {
		return CreateProfileOutput{}, err
	}
	if err := domain.ValidateName("last_name", in.LastName); err != nil {
		return CreateProfileOutput{}, err
	}

	switch in.Role {
	case domain.RoleClient:
		if in.Freelancer != nil {
			return CreateProfileOutput{}, fmt.Errorf("freelancer details not allowed for client")
		}
		if in.Client != nil && in.Client.VerificationStatus == "" {
			in.Client.VerificationStatus = domain.VerificationStatusPending
		}
	case domain.RoleFreelancer:
		if in.Client != nil {
			return CreateProfileOutput{}, fmt.Errorf("client details not allowed for freelancer")
		}
		if in.Freelancer != nil {
			if in.Freelancer.VerificationStatus == "" {
				in.Freelancer.VerificationStatus = domain.VerificationStatusPending
			}
			if in.Freelancer.Availability == "" {
				in.Freelancer.Availability = domain.AvailabilityAsNeeded
			}
		}
	case domain.RoleAdmin:
		if in.Client != nil || in.Freelancer != nil {
			return CreateProfileOutput{}, fmt.Errorf("role details not allowed for admin")
		}
	}

	displayName := strings.TrimSpace(in.DisplayName)
	if displayName == "" {
		displayName = domain.BuildDisplayName(in.FirstName, in.LastName)
	}
	if displayName == "" {
		return CreateProfileOutput{}, fmt.Errorf("display_name is required")
	}

	now := time.Now().UTC()
	if uc.Clock != nil {
		now = uc.Clock.Now()
	}

	profile := domain.Profile{
		UserID:        in.UserID,
		Role:          in.Role,
		FirstName:     in.FirstName,
		LastName:      in.LastName,
		DisplayName:   displayName,
		AvatarURL:     in.AvatarURL,
		Language:      "en",
		ContactEmail:  in.ContactEmail,
		TaxID:         "",
		VerificationStatus: "",
		AccountStatus: domain.AccountStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if in.Client != nil {
		profile.TaxID = strings.TrimSpace(in.Client.TaxID)
		profile.VerificationStatus = strings.TrimSpace(in.Client.VerificationStatus)
	}
	if in.Freelancer != nil {
		profile.VerificationStatus = strings.TrimSpace(in.Freelancer.VerificationStatus)
	}

	profileID, err := uc.Profiles.Create(ctx, profile, in.Client, in.Freelancer)
	if err != nil {
		return CreateProfileOutput{}, err
	}

	return CreateProfileOutput{ProfileID: profileID}, nil
}
