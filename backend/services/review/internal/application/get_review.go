package application

import (
	"context"

	"jobconnect/review/internal/domain"
)

type GetReview struct {
	Reviews ReviewRepository
}

type GetReviewInput struct {
	ReviewID int64
}

type GetReviewOutput struct {
	Review domain.Review
}

func (uc *GetReview) Execute(ctx context.Context, in GetReviewInput) (GetReviewOutput, error) {
	r, err := uc.Reviews.GetByID(ctx, in.ReviewID)
	if err != nil {
		return GetReviewOutput{}, err
	}
	return GetReviewOutput{Review: r}, nil
}
