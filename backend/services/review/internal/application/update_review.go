package application

import (
	"context"
	"fmt"

	"jobconnect/review/internal/domain"

	"github.com/google/uuid"
)

type UpdateReview struct {
	Reviews ReviewRepository
	Clock   Clock
	Events  ReviewEventsPublisher
}

type UpdateReviewInput struct {
	ReviewID    int64
	RequesterID uuid.UUID
	Rating      *int32
	Title       *string
	Comment     *string
}

type UpdateReviewOutput struct {
	Review domain.Review
}

func (uc *UpdateReview) Execute(ctx context.Context, in UpdateReviewInput) (UpdateReviewOutput, error) {
	r, err := uc.Reviews.GetByID(ctx, in.ReviewID)
	if err != nil {
		return UpdateReviewOutput{}, err
	}

	if r.ReviewerID != in.RequesterID {
		return UpdateReviewOutput{}, fmt.Errorf("you can only update your own review")
	}

	if !r.IsWithinGracePeriod(uc.Clock.Now()) {
		return UpdateReviewOutput{}, fmt.Errorf("the 24-hour grace period for updating this review has expired")
	}

	if in.Rating != nil {
		r.Rating = *in.Rating
	}
	if in.Title != nil {
		r.Title = *in.Title
	}
	if in.Comment != nil {
		r.Comment = *in.Comment
	}

	if err := domain.ValidateCreate(r); err != nil {
		return UpdateReviewOutput{}, err
	}

	now := uc.Clock.Now()
	r.UpdatedAt = &now

	updated, err := uc.Reviews.Update(ctx, r)
	if err != nil {
		return UpdateReviewOutput{}, err
	}
	if uc.Events != nil {
		_ = uc.Events.PublishReviewUpdated(ctx, updated)
	}

	return UpdateReviewOutput{Review: updated}, nil
}
