package applications

import (
	"context"
	"fmt"
	"jobconnect/reviews/internal/domain"
)

type UpdateReview struct {
	Reviews ReviewRepository
	Clock   Clock
}

type UpdateReviewInput struct {
	ID      int64
	Rating  int
	Title   string
	Comment string
	UserID  string
}

type UpdateReviewOutput struct {
	Review domain.Review
}

func (uc *UpdateReview) Execute(ctx context.Context, input UpdateReviewInput) (UpdateReviewOutput, error) {
	review, err := uc.Reviews.GetByID(ctx, input.ID)
	if err != nil {
		return UpdateReviewOutput{}, err
	}

	if !uc.canEdit(review, input.UserID) {
		return UpdateReviewOutput{}, fmt.Errorf("user %s is not authorized to edit this review", input.UserID)
	}

	review.Rating = input.Rating
	review.Title = input.Title
	review.Comment = input.Comment
	review.UpdatedAt = uc.Clock.Now()

	updated, err := uc.Reviews.Update(ctx, review)
	if err != nil {
		return UpdateReviewOutput{}, err
	}

	return UpdateReviewOutput{Review: updated}, nil
}

func (uc *UpdateReview) canEdit(r domain.Review, userID string) bool {
	if r.ReviewerRole == domain.RoleClient {
		return r.ClientID == userID
	}
	return r.FreelancerID == userID
}
