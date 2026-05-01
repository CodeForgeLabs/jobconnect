package applications

import (
	"context"
)

type GetUserRatingSummary struct {
	Reviews ReviewRepository
}

type GetUserRatingSummaryInput struct {
	UserID string
}

type GetUserRatingSummaryOutput struct {
	AverageRating float64
	TotalReviews  int64
}

func (uc *GetUserRatingSummary) Execute(
	ctx context.Context,
	input GetUserRatingSummaryInput,
) (GetUserRatingSummaryOutput, error) {

	avg, total, err := uc.Reviews.GetUserRatingSummary(ctx, input.UserID)
	if err != nil {
		return GetUserRatingSummaryOutput{}, err
	}

	return GetUserRatingSummaryOutput{
		AverageRating: avg,
		TotalReviews:  total,
	}, nil
}
