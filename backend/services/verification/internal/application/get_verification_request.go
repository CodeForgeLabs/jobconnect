package application

import (
	"context"
	"fmt"

	"jobconnect/verification/internal/domain"
)

type GetVerificationRequestInput struct {
	RequestID int64
}

type GetVerificationRequest struct {
	Repo VerificationRepository
}

func (uc *GetVerificationRequest) Execute(ctx context.Context, in GetVerificationRequestInput) (domain.VerificationRequest, error) {
	if in.RequestID <= 0 {
		return domain.VerificationRequest{}, fmt.Errorf("request_id is required")
	}
	return uc.Repo.GetByID(ctx, in.RequestID)
}
