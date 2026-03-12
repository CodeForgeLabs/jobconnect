package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type ListMyContracts struct {
	Contracts ContractRepository
}

type ListMyContractsInput struct {
	ActorID   uuid.UUID
	Status    string
	PageSize  int32
	PageToken string
}

type ListMyContractsOutput struct {
	Contracts     []domain.Contract
	NextPageToken string
}

func (uc *ListMyContracts) Execute(ctx context.Context, in ListMyContractsInput) (ListMyContractsOutput, error) {
	if uc.Contracts == nil {
		return ListMyContractsOutput{}, fmt.Errorf("list contracts dependencies are not configured")
	}
	if in.ActorID == uuid.Nil {
		return ListMyContractsOutput{}, fmt.Errorf("actor_id is required")
	}
	status := strings.ToLower(strings.TrimSpace(in.Status))
	if status != "" {
		if err := domain.ValidateStatus(status); err != nil {
			return ListMyContractsOutput{}, err
		}
	}

	pageSize := int(in.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := 0
	if strings.TrimSpace(in.PageToken) != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(in.PageToken))
		if err != nil || parsed < 0 {
			return ListMyContractsOutput{}, fmt.Errorf("invalid page_token")
		}
		offset = parsed
	}

	items, err := uc.Contracts.ListByActor(ctx, in.ActorID, status, pageSize, offset)
	if err != nil {
		return ListMyContractsOutput{}, err
	}

	next := ""
	if len(items) == pageSize {
		next = strconv.Itoa(offset + len(items))
	}

	return ListMyContractsOutput{Contracts: items, NextPageToken: next}, nil
}
