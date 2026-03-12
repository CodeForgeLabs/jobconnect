package application

import (
	"context"
	"fmt"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type GetContract struct {
	Contracts ContractRepository
}

type GetContractInput struct {
	ContractID int64
	ActorID    uuid.UUID
}

type GetContractOutput struct {
	Contract domain.Contract
}

func (uc *GetContract) Execute(ctx context.Context, in GetContractInput) (GetContractOutput, error) {
	if uc.Contracts == nil {
		return GetContractOutput{}, fmt.Errorf("get contract dependencies are not configured")
	}
	if in.ContractID <= 0 {
		return GetContractOutput{}, fmt.Errorf("contract_id is required")
	}
	if in.ActorID == uuid.Nil {
		return GetContractOutput{}, fmt.Errorf("actor_id is required")
	}
	c, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ActorID)
	if err != nil {
		return GetContractOutput{}, err
	}
	return GetContractOutput{Contract: c}, nil
}
