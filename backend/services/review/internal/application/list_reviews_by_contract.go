package application

import (
	"context"

	"jobconnect/review/internal/domain"
)

type ListReviewsByContract struct {
	Reviews ReviewRepository
}

type ListReviewsByContractInput struct {
	ContractID int64
}

type ListReviewsByContractOutput struct {
	Reviews []domain.Review
}

func (uc *ListReviewsByContract) Execute(ctx context.Context, in ListReviewsByContractInput) (ListReviewsByContractOutput, error) {
	reviews, err := uc.Reviews.ListByContract(ctx, in.ContractID)
	if err != nil {
		return ListReviewsByContractOutput{}, err
	}
	return ListReviewsByContractOutput{Reviews: reviews}, nil
}
