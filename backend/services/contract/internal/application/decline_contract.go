package application

import (
	"context"
	"fmt"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type DeclineContract struct {
	Contracts ContractRepository
	Proposals ProposalSync
	Clock     Clock
}

type DeclineContractInput struct {
	ContractID   int64
	FreelancerID uuid.UUID
	Reason       string
}

type DeclineContractOutput struct {
	Contract domain.Contract
}

func (uc *DeclineContract) Execute(ctx context.Context, in DeclineContractInput) (DeclineContractOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil || uc.Proposals == nil {
		return DeclineContractOutput{}, fmt.Errorf("decline contract dependencies are not configured")
	}
	if in.ContractID <= 0 {
		return DeclineContractOutput{}, fmt.Errorf("contract_id is required")
	}
	if in.FreelancerID == uuid.Nil {
		return DeclineContractOutput{}, fmt.Errorf("freelancer_id is required")
	}
	current, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.FreelancerID)
	if err != nil {
		return DeclineContractOutput{}, err
	}
	if !domain.CanDecline(current.Status) {
		return DeclineContractOutput{}, fmt.Errorf("contract is not in a decline-able state")
	}
	now := uc.Clock.Now()
	if err := uc.Contracts.SetStatusForFreelancer(ctx, in.ContractID, in.FreelancerID, domain.StatusDeclined, now); err != nil {
		return DeclineContractOutput{}, err
	}
	if current.ProposalID > 0 {
		if err := uc.Proposals.ReleaseHired(ctx, current.ProposalID, current.ClientID, in.Reason); err != nil {
			_ = uc.Contracts.SetStatusForFreelancer(ctx, in.ContractID, in.FreelancerID, domain.StatusPendingAcceptance, uc.Clock.Now())
			return DeclineContractOutput{}, fmt.Errorf("sync proposal status: %w", err)
		}
	}
	_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{
		ContractID: in.ContractID,
		Status:     domain.StatusDeclined,
		Reason:     in.Reason,
		ActorID:    in.FreelancerID,
		CreatedAt:  now,
	})
	updated, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.FreelancerID)
	if err != nil {
		return DeclineContractOutput{}, err
	}
	return DeclineContractOutput{Contract: updated}, nil
}
