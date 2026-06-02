package application

import (
	"context"

	"jobconnect/verification/internal/domain"
)

type ListPendingVerificationsInput struct {
	PageSize int32
	Page     int32
}

type ListPendingVerifications struct {
	Repo VerificationRepository
}

func (uc *ListPendingVerifications) Execute(ctx context.Context, in ListPendingVerificationsInput) ([]domain.VerificationRequest, error) {
	pageSize := in.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	page := in.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	return uc.Repo.ListPending(ctx, pageSize, offset)
}
