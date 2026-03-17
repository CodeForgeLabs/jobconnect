package application

import (
	"context"

	"github.com/google/uuid"
)

type GetUserRatingSummary struct {
	Reviews ReviewRepository
}

type GetUserRatingSummaryInput struct {
	UserID uuid.UUID
}

type GetUserRatingSummaryOutput struct {
	AverageRating float64
	TotalReviews  int64
}

func (uc *GetUserRatingSummary) Execute(ctx context.Context, in GetUserRatingSummaryInput) (GetUserRatingSummaryOutput, error) {
	avg, count, err := uc.Reviews.GetRatingSummary(ctx, in.UserID)
	if err != nil {
		return GetUserRatingSummaryOutput{}, err
	}
	return GetUserRatingSummaryOutput{AverageRating: avg, TotalReviews: count}, nil
}
