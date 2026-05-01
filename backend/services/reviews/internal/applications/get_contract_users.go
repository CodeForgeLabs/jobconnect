package applications

import "context"

type GetContractUsers struct {
	Reviews ReviewRepository
	Clock   Clock
}

type GetContractUsersInput struct {
	ContractID int64
}

type GetContractUsersOutput struct {
	ClientID     string
	FreelancerID string
}

func (uc *GetContractUsers) Execute(ctx context.Context, input GetContractUsersInput) (GetContractUsersOutput, error) {
	review, err := uc.Reviews.GetContractUsers(ctx, input.ContractID)
	if err != nil {
		return GetContractUsersOutput{}, err
	}

	return GetContractUsersOutput{
		ClientID:     review.ClientID,
		FreelancerID: review.FreelancerID,
	}, nil
}
