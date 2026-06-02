package applications

import (
	"context"
	"jobconnect/reviews/internal/domain"
)

type CreateReview struct {
	Reviews ReviewRepository
	Clock   Clock
}

type CreateReviewInput struct {
	ContractID   int64
	ClientID     string
	FreelancerID string
	ReviewerRole domain.ReviewerRole
	Rating       int
	Title        string
	Comment      string
}

type CreateReviewOutput struct {
	Review domain.Review
}

func (uc *CreateReview) Execute(ctx context.Context, input CreateReviewInput) (CreateReviewOutput, error) {
	now := uc.Clock.Now()

	review := domain.Review{
		ContractID:   input.ContractID,
		ClientID:     input.ClientID,
		FreelancerID: input.FreelancerID,
		ReviewerRole: input.ReviewerRole,
		Rating:       input.Rating,
		Title:        input.Title,
		Comment:      input.Comment,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	created, err := uc.Reviews.Create(ctx, review)
	if err != nil {
		return CreateReviewOutput{}, err
	}

	return CreateReviewOutput{Review: created}, nil
}
