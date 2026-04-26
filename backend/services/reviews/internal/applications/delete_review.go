package applications

import (
	"context"
	"fmt"
	"jobconnect/reviews/internal/domain"
)

type DeleteReview struct {
	Reviews ReviewRepository
}

type DeleteReviewInput struct {
	ID     int64
	UserID string
}

type DeleteReviewOutput struct {
	Success bool
}

func (uc *DeleteReview) Execute(ctx context.Context, input DeleteReviewInput) (DeleteReviewOutput, error) {
	review, err := uc.Reviews.GetByID(ctx, input.ID)
	if err != nil {
		return DeleteReviewOutput{}, err
	}

	if !uc.canDelete(review, input.UserID) {
		return DeleteReviewOutput{}, fmt.Errorf("user %s is not authorized to delete this review", input.UserID)
	}

	err = uc.Reviews.Delete(ctx, input.ID)
	if err != nil {
		return DeleteReviewOutput{}, err
	}

	return DeleteReviewOutput{Success: true}, nil
}

func (uc *DeleteReview) canDelete(r domain.Review, userID string) bool {
	if r.ReviewerRole == domain.RoleClient {
		return r.ClientID == userID
	}
	return r.FreelancerID == userID
}
