package application

import (
	"context"
	"fmt"

	"jobconnect/review/internal/domain"

	"github.com/google/uuid"
)

type ReplyToReview struct {
	Reviews ReviewRepository
	Clock   Clock
}

type ReplyToReviewInput struct {
	ReviewID     int64
	RequesterID  uuid.UUID
	ReplyComment string
}

type ReplyToReviewOutput struct {
	Review domain.Review
}

func (uc *ReplyToReview) Execute(ctx context.Context, in ReplyToReviewInput) (ReplyToReviewOutput, error) {
	r, err := uc.Reviews.GetByID(ctx, in.ReviewID)
	if err != nil {
		return ReplyToReviewOutput{}, err
	}

	if r.RevieweeID != in.RequesterID {
		return ReplyToReviewOutput{}, fmt.Errorf("you can only reply to reviews about yourself")
	}

	if r.ReplyComment != nil {
		return ReplyToReviewOutput{}, fmt.Errorf("you have already replied to this review")
	}

	if err := domain.ValidateReply(in.ReplyComment); err != nil {
		return ReplyToReviewOutput{}, err
	}

	now := uc.Clock.Now()
	r.ReplyComment = &in.ReplyComment
	r.RepliedAt = &now

	updated, err := uc.Reviews.Update(ctx, r)
	if err != nil {
		return ReplyToReviewOutput{}, err
	}

	return ReplyToReviewOutput{Review: updated}, nil
}
