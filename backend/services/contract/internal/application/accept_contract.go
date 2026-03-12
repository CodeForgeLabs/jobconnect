package application

import (
	"context"
	"fmt"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type AcceptContract struct {
	Contracts ContractRepository
	Proposals ProposalStatusSync
	Jobs      JobStatusSync
	Clock     Clock
}

type AcceptContractInput struct {
	ContractID   int64
	FreelancerID uuid.UUID
}

type AcceptContractOutput struct {
	Contract domain.Contract
}

func (uc *AcceptContract) Execute(ctx context.Context, in AcceptContractInput) (AcceptContractOutput, error) {
	if uc.Contracts == nil || uc.Clock == nil {
		return AcceptContractOutput{}, fmt.Errorf("accept contract dependencies are not configured")
	}
	if in.ContractID <= 0 {
		return AcceptContractOutput{}, fmt.Errorf("contract_id is required")
	}
	if in.FreelancerID == uuid.Nil {
		return AcceptContractOutput{}, fmt.Errorf("freelancer_id is required")
	}
	current, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.FreelancerID)
	if err != nil {
		return AcceptContractOutput{}, err
	}
	if !domain.CanAccept(current.Status) {
		return AcceptContractOutput{}, fmt.Errorf("contract is not in an acceptable state")
	}
	activatedAt := uc.Clock.Now()
	if err := uc.Contracts.SetStatusForFreelancer(ctx, in.ContractID, in.FreelancerID, domain.StatusActive, activatedAt); err != nil {
		return AcceptContractOutput{}, err
	}
	updated, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.FreelancerID)
	if err != nil {
		return AcceptContractOutput{}, err
	}

	if updated.ProposalID > 0 && uc.Proposals != nil {
		err = uc.Proposals.SetHired(ctx, updated.ProposalID, updated.ClientID, "contract accepted")
		if err != nil {
			_ = uc.Contracts.SetStatusForFreelancer(ctx, in.ContractID, in.FreelancerID, domain.StatusPendingAcceptance, uc.Clock.Now())
			return AcceptContractOutput{}, fmt.Errorf("sync proposal status: %w", err)
		}
	}
	if updated.JobID > 0 && uc.Jobs != nil {
		err = uc.Jobs.SetInProgress(ctx, updated.JobID, updated.ClientID)
		if err != nil {
			_ = uc.Contracts.SetStatusForFreelancer(ctx, in.ContractID, in.FreelancerID, domain.StatusPendingAcceptance, uc.Clock.Now())
			return AcceptContractOutput{}, fmt.Errorf("sync job status: %w", err)
		}
	}
	_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{
		ContractID: in.ContractID,
		Status:     domain.StatusActive,
		Reason:     "accepted by freelancer",
		ActorID:    in.FreelancerID,
		CreatedAt:  activatedAt,
	})

	return AcceptContractOutput{Contract: updated}, nil
}
