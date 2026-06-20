package applications

import (
	"context"

	"jobconnect/reviews/internal/domain"
)

type GetReview struct {
	Reviews ReviewRepository
}

type GetReviewInput struct {
	ID int64
}

type GetReviewOutput struct {
	Review domain.Review
}

func (uc *GetReview) Execute(ctx context.Context, input GetReviewInput) (GetReviewOutput, error) {
	review, err := uc.Reviews.GetByID(ctx, input.ID)
	if err != nil {
		return GetReviewOutput{}, err
	}

	return GetReviewOutput{Review: review}, nil
}
