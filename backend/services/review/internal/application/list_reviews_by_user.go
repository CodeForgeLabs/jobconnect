package application

import (
	"context"
	"fmt"
	"strconv"

	"jobconnect/review/internal/domain"

	"github.com/google/uuid"
)

type ListReviewsByUser struct {
	Reviews ReviewRepository
}

type ListReviewsByUserInput struct {
	UserID    uuid.UUID
	PageSize  int32
	PageToken string
}

type ListReviewsByUserOutput struct {
	Reviews       []domain.Review
	NextPageToken string
}

func (uc *ListReviewsByUser) Execute(ctx context.Context, in ListReviewsByUserInput) (ListReviewsByUserOutput, error) {
	limit := int(in.PageSize)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := 0
	if in.PageToken != "" {
		parsed, err := strconv.Atoi(in.PageToken)
		if err != nil {
			return ListReviewsByUserOutput{}, fmt.Errorf("invalid page_token")
		}
		offset = parsed
	}

	reviews, err := uc.Reviews.ListByReviewee(ctx, in.UserID, limit+1, offset)
	if err != nil {
		return ListReviewsByUserOutput{}, err
	}

	var next string
	if len(reviews) > limit {
		reviews = reviews[:limit]
		next = strconv.Itoa(offset + limit)
	}
	return ListReviewsByUserOutput{Reviews: reviews, NextPageToken: next}, nil
}
