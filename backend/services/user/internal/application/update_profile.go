package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type UpdateProfileInput struct {
	UserID       uuid.UUID
	DisplayName  *string
	AvatarURL    *string
	ContactEmail *string
	ContactPhone *string
	Bio          *string
	CompanyName  *string
	TaxID        *string
	Headline     *string
	Skills       []string
	HourlyRate   *float64
	Availability *string
	Location     *string
}

type UpdateProfileOutput struct {
	Profile      domain.Profile
	Client       *domain.ClientProfile
	Freelancer   *domain.FreelancerProfile
	Completeness uint32
	Missing      []string
}

type UpdateProfile struct {
	Profiles ProfileRepository
	Clock    Clock
}

func (uc *UpdateProfile) Execute(ctx context.Context, in UpdateProfileInput) (UpdateProfileOutput, error) {
	if in.UserID == uuid.Nil {
		return UpdateProfileOutput{}, fmt.Errorf("user_id is required")
	}

	profile, client, freelancer, err := uc.Profiles.GetByUserID(ctx, in.UserID)
	if err != nil {
		return UpdateProfileOutput{}, err
	}

	if in.DisplayName != nil {
		if err := domain.ValidateDisplayName(*in.DisplayName); err != nil {
			return UpdateProfileOutput{}, err
		}
		profile.DisplayName = strings.TrimSpace(*in.DisplayName)
	}
	if in.ContactEmail != nil {
		profile.ContactEmail = strings.TrimSpace(*in.ContactEmail)
	}
	if in.ContactPhone != nil {
		profile.ContactPhone = strings.TrimSpace(*in.ContactPhone)
	}
	if in.Bio != nil {
		profile.Bio = strings.TrimSpace(*in.Bio)
	}
	if in.Location != nil {
		profile.Location = strings.TrimSpace(*in.Location)
	}
	if in.TaxID != nil {
		profile.TaxID = strings.TrimSpace(*in.TaxID)
	}
	if in.AvatarURL != nil {
		profile.AvatarURL = strings.TrimSpace(*in.AvatarURL)
	}

	switch profile.Role {
	case domain.RoleClient:
		if client == nil {
			client = &domain.ClientProfile{}
		}
		if in.CompanyName != nil {
			client.CompanyName = strings.TrimSpace(*in.CompanyName)
		}
		if in.Headline != nil || in.Skills != nil || in.HourlyRate != nil || in.Availability != nil {
			return UpdateProfileOutput{}, fmt.Errorf("freelancer fields are not allowed for client")
		}
	case domain.RoleFreelancer:
		if freelancer == nil {
			freelancer = &domain.FreelancerProfile{}
		}
		if in.Headline != nil {
			freelancer.Headline = strings.TrimSpace(*in.Headline)
		}
		if in.Skills != nil {
			skills := make([]string, 0, len(in.Skills))
			for _, item := range in.Skills {
				trimmed := strings.TrimSpace(item)
				if trimmed != "" {
					skills = append(skills, trimmed)
				}
			}
			freelancer.Skills = skills
		}
		if in.HourlyRate != nil {
			if *in.HourlyRate < 0 {
				return UpdateProfileOutput{}, fmt.Errorf("hourly_rate must be greater than or equal to 0")
			}
			freelancer.HourlyRate = *in.HourlyRate
		}
		if in.Availability != nil {
			freelancer.Availability = strings.TrimSpace(*in.Availability)
		}
		if in.CompanyName != nil {
			return UpdateProfileOutput{}, fmt.Errorf("client fields are not allowed for freelancer")
		}
	case domain.RoleAdmin:
		if in.CompanyName != nil || in.TaxID != nil || in.Headline != nil || in.Skills != nil || in.HourlyRate != nil || in.Availability != nil {
			return UpdateProfileOutput{}, fmt.Errorf("role-specific fields are not allowed for admin")
		}
	}

	if err := uc.Profiles.Update(ctx, profile, client, freelancer); err != nil {
		return UpdateProfileOutput{}, err
	}

	completeness, missing := computeCompleteness(profile, client, freelancer, readinessSignals{})
	return UpdateProfileOutput{
		Profile:      profile,
		Client:       client,
		Freelancer:   freelancer,
		Completeness: completeness,
		Missing:      missing,
	}, nil
}
