package application

import (
	"context"
	"fmt"

	"jobconnect/verification/internal/domain"

	"github.com/google/uuid"
)

type GetMyVerificationStatusInput struct {
	UserID uuid.UUID
}

type GetMyVerificationStatus struct {
	Repo VerificationRepository
}

func (uc *GetMyVerificationStatus) Execute(ctx context.Context, in GetMyVerificationStatusInput) (domain.VerificationRequest, error) {
	if in.UserID == uuid.Nil {
		return domain.VerificationRequest{}, fmt.Errorf("user_id is required")
	}
	return uc.Repo.GetLatestByUserID(ctx, in.UserID)
}
