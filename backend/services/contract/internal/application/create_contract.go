package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type CreateContract struct {
	Contracts ContractRepository
	Clock     Clock
}

type CreateContractInput struct {
	ClientID        uuid.UUID
	FreelancerID    uuid.UUID
	JobID           int64
	ProposalID      int64
	ContractType    string
	Title           string
	Description     string
	Currency        string
	HourlyRate      float64
	FixedTotal      float64
	WeeklyHourLimit int32
	Milestones      []domain.Milestone
}

type CreateContractOutput struct {
	Contract domain.Contract
}

func (uc *CreateContract) Execute(ctx context.Context, in CreateContractInput) (CreateContractOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return CreateContractOutput{}, fmt.Errorf("create contract dependencies are not configured")
	}
	now := uc.Clock.Now()
	c := domain.Contract{
		ClientID:        in.ClientID,
		FreelancerID:    in.FreelancerID,
		JobID:           in.JobID,
		ProposalID:      in.ProposalID,
		ContractType:    strings.ToLower(strings.TrimSpace(in.ContractType)),
		Status:          domain.StatusPendingAcceptance,
		Title:           strings.TrimSpace(in.Title),
		Description:     strings.TrimSpace(in.Description),
		Currency:        strings.ToUpper(strings.TrimSpace(in.Currency)),
		HourlyRate:      in.HourlyRate,
		FixedTotal:      in.FixedTotal,
		WeeklyHourLimit: in.WeeklyHourLimit,
		Milestones:      in.Milestones,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := domain.ValidateForCreate(c); err != nil {
		return CreateContractOutput{}, err
	}

	id, err := uc.Contracts.Create(ctx, c)
	if err != nil {
		return CreateContractOutput{}, err
	}
	_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{
		ContractID: id,
		Status:     domain.StatusPendingAcceptance,
		Reason:     "contract created",
		ActorID:    in.ClientID,
		CreatedAt:  now,
	})
	persisted, err := uc.Contracts.GetByID(ctx, id)
	if err != nil {
		return CreateContractOutput{}, err
	}
	return CreateContractOutput{Contract: persisted}, nil
}
