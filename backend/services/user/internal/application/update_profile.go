package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type UpdateProfileInput struct {
	UserID           uuid.UUID
	DisplayName      *string
	AvatarURL        *string
	Language         *string
	ContactEmail     *string
	ContactPhone     *string
	Bio              *string
	FirstName        *string
	LastName         *string
	CompanyName      *string
	BillingAddress   *string
	TaxID            *string
	Headline         *string
	Skills           []string
	ExperienceLevel  *string
	HourlyRate       *float64
	Availability     *string
	Location         *string
	LastActiveAtUnix *int64
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
	if in.Language != nil {
		profile.Language = strings.TrimSpace(*in.Language)
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
	if in.FirstName != nil {
		if err := domain.ValidateOptionalName("first_name", *in.FirstName); err != nil {
			return UpdateProfileOutput{}, err
		}
		profile.FirstName = strings.TrimSpace(*in.FirstName)
	}
	if in.LastName != nil {
		if err := domain.ValidateOptionalName("last_name", *in.LastName); err != nil {
			return UpdateProfileOutput{}, err
		}
		profile.LastName = strings.TrimSpace(*in.LastName)
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
		if in.BillingAddress != nil {
			client.BillingAddress = strings.TrimSpace(*in.BillingAddress)
		}
		if in.TaxID != nil {
			client.TaxID = strings.TrimSpace(*in.TaxID)
		}
		if in.Headline != nil || in.ExperienceLevel != nil || len(in.Skills) > 0 || in.HourlyRate != nil || in.Availability != nil || in.Location != nil || in.LastActiveAtUnix != nil {
			return UpdateProfileOutput{}, fmt.Errorf("freelancer fields are not allowed for client")
		}
	case domain.RoleFreelancer:
		if freelancer == nil {
			freelancer = &domain.FreelancerProfile{}
		}
		if in.Headline != nil {
			freelancer.Headline = strings.TrimSpace(*in.Headline)
		}
		if in.ExperienceLevel != nil {
			freelancer.ExperienceLevel = strings.TrimSpace(*in.ExperienceLevel)
		}
		if len(in.Skills) > 0 {
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
		if in.Location != nil {
			freelancer.Location = strings.TrimSpace(*in.Location)
		}
		if in.LastActiveAtUnix != nil {
			if *in.LastActiveAtUnix <= 0 {
				freelancer.LastActiveAt = nil
			} else {
				t := time.Unix(*in.LastActiveAtUnix, 0).UTC()
				freelancer.LastActiveAt = &t
			}
		}
		if in.CompanyName != nil || in.BillingAddress != nil || in.TaxID != nil {
			return UpdateProfileOutput{}, fmt.Errorf("client fields are not allowed for freelancer")
		}
	case domain.RoleAdmin:
		if in.CompanyName != nil || in.BillingAddress != nil || in.TaxID != nil || in.Headline != nil || in.ExperienceLevel != nil || len(in.Skills) > 0 || in.HourlyRate != nil || in.Availability != nil || in.Location != nil || in.LastActiveAtUnix != nil {
			return UpdateProfileOutput{}, fmt.Errorf("role-specific fields are not allowed for admin")
		}
	}

	if err := uc.Profiles.Update(ctx, profile, client, freelancer); err != nil {
		return UpdateProfileOutput{}, err
	}

	completeness, missing := computeCompleteness(profile, client, freelancer)
	return UpdateProfileOutput{
		Profile:      profile,
		Client:       client,
		Freelancer:   freelancer,
		Completeness: completeness,
		Missing:      missing,
	}, nil
}
