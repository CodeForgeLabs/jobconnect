package applications

import (
	"context"
	"jobconnect/reviews/internal/domain"
)

type ListReviews struct {
	Reviews ReviewRepository
}

type ListReviewsInput struct {
	UserID string
	Role   domain.ReviewerRole
	Limit  int
	Offset int
}

type ListReviewsOutput struct {
	Reviews []domain.Review
}

func (uc *ListReviews) Execute(ctx context.Context, input ListReviewsInput) (ListReviewsOutput, error) {
	reviews, err := uc.Reviews.ListByUser(ctx, input.UserID, input.Role, input.Limit, input.Offset)
	if err != nil {
		return ListReviewsOutput{}, err
	}

	return ListReviewsOutput{Reviews: reviews}, nil
}
