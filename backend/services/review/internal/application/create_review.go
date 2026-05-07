package application

import (
	"context"
	"fmt"

	"jobconnect/review/internal/domain"

	"github.com/google/uuid"
)

type CreateReview struct {
	Reviews ReviewRepository
	Clock   Clock
	Events  ReviewEventsPublisher
}

type CreateReviewInput struct {
	ContractID   int64
	ReviewerID   uuid.UUID
	ReviewerRole string
	RevieweeID   uuid.UUID
	Rating       int32
	Title        string
	Comment      string
}

type CreateReviewOutput struct {
	Review domain.Review
}

func (uc *CreateReview) Execute(ctx context.Context, in CreateReviewInput) (CreateReviewOutput, error) {
	r := domain.Review{
		ContractID:   in.ContractID,
		ReviewerID:   in.ReviewerID,
		RevieweeID:   in.RevieweeID,
		ReviewerRole: in.ReviewerRole,
		Rating:       in.Rating,
		Title:        in.Title,
		Comment:      in.Comment,
		CreatedAt:    uc.Clock.Now(),
	}
	if err := domain.ValidateCreate(r); err != nil {
		return CreateReviewOutput{}, err
	}

	exists, err := uc.Reviews.ExistsByContractAndReviewer(ctx, in.ContractID, in.ReviewerID)
	if err != nil {
		return CreateReviewOutput{}, err
	}
	if exists {
		return CreateReviewOutput{}, fmt.Errorf("you have already reviewed this contract")
	}

	id, err := uc.Reviews.Create(ctx, r)
	if err != nil {
		return CreateReviewOutput{}, err
	}
	r.ID = id
	if uc.Events != nil {
		_ = uc.Events.PublishReviewCreated(ctx, r)
	}
	return CreateReviewOutput{Review: r}, nil
}
