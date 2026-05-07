package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type DeleteReview struct {
	Reviews ReviewRepository
	Clock   Clock
	Events  ReviewEventsPublisher
}

type DeleteReviewInput struct {
	ReviewID    int64
	RequesterID uuid.UUID
}

func (uc *DeleteReview) Execute(ctx context.Context, in DeleteReviewInput) error {
	r, err := uc.Reviews.GetByID(ctx, in.ReviewID)
	if err != nil {
		return err
	}

	if r.ReviewerID != in.RequesterID {
		return fmt.Errorf("you can only delete your own review")
	}

	if !r.IsWithinGracePeriod(uc.Clock.Now()) {
		return fmt.Errorf("the 24-hour grace period for deleting this review has expired")
	}

	if err := uc.Reviews.Delete(ctx, in.ReviewID); err != nil {
		return err
	}
	if uc.Events != nil {
		_ = uc.Events.PublishReviewDeleted(ctx, in.ReviewID, r.RevieweeID.String())
	}
	return nil
}
