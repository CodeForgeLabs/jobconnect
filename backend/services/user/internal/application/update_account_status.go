package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type UpdateAccountStatusInput struct {
	UserID           uuid.UUID
	Status           string
	SuspensionReason *string
	Visibility       string
}

type UpdateAccountStatusOutput struct {
	Profile    domain.Profile
	Client     *domain.ClientProfile
	Freelancer *domain.FreelancerProfile
}

type UpdateAccountStatus struct {
	Profiles ProfileRepository
	Clock    Clock
}

func (uc *UpdateAccountStatus) Execute(ctx context.Context, in UpdateAccountStatusInput) (UpdateAccountStatusOutput, error) {
	if in.UserID == uuid.Nil {
		return UpdateAccountStatusOutput{}, fmt.Errorf("user_id is required")
	}
	status := strings.ToUpper(strings.TrimSpace(in.Status))
	status = strings.TrimPrefix(status, "ACCOUNT_STATUS_")
	if err := domain.ValidateAccountStatus(status); err != nil {
		return UpdateAccountStatusOutput{}, err
	}
	visibility := strings.ToUpper(strings.TrimSpace(in.Visibility))
	visibility = strings.TrimPrefix(visibility, "PROFILE_VISIBILITY_")
	if err := domain.ValidateProfileVisibility(visibility); err != nil {
		return UpdateAccountStatusOutput{}, err
	}

	reason := ""
	if in.SuspensionReason != nil {
		reason = strings.TrimSpace(*in.SuspensionReason)
	}
	if status == domain.AccountStatusSuspended && reason == "" {
		return UpdateAccountStatusOutput{}, fmt.Errorf("suspension_reason is required when status is SUSPENDED")
	}
	if status != domain.AccountStatusSuspended {
		reason = ""
	}

	now := uc.Clock.Now()
	profile, client, freelancer, err := uc.Profiles.UpdateAccountState(ctx, in.UserID, status, reason, visibility, now)
	if err != nil {
		return UpdateAccountStatusOutput{}, err
	}
	return UpdateAccountStatusOutput{Profile: profile, Client: client, Freelancer: freelancer}, nil
}
