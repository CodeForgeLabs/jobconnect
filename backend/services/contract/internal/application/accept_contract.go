package application

import (
	"context"
	"fmt"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type AcceptContract struct {
	Contracts ContractRepository
	Proposals ProposalSync
	Jobs      JobStatusSync
	Actors    ActorPolicy
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
	if uc.Contracts == nil || uc.Proposals == nil || uc.Clock == nil || uc.Actors == nil {
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
	if err := uc.Actors.EnsureFreelancerCanWork(ctx, in.FreelancerID); err != nil {
		return AcceptContractOutput{}, err
	}
	if err := uc.Actors.EnsureClientCanHire(ctx, current.ClientID); err != nil {
		return AcceptContractOutput{}, err
	}
	activatedAt := uc.Clock.Now()
	if err := uc.Contracts.SetStatusForFreelancer(ctx, in.ContractID, in.FreelancerID, domain.StatusActive, activatedAt); err != nil {
		return AcceptContractOutput{}, err
	}
	updated, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.FreelancerID)
	if err != nil {
		return AcceptContractOutput{}, err
	}

	if updated.ProposalID > 0 {
		if err := uc.Proposals.SetHired(ctx, updated.ProposalID, updated.ClientID, "accepted by freelancer"); err != nil {
			revertErr := uc.Contracts.SetStatusForFreelancer(ctx, in.ContractID, in.FreelancerID, domain.StatusPendingAcceptance, uc.Clock.Now())
			_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{
				ContractID: in.ContractID,
				Status:     domain.StatusPendingAcceptance,
				Reason:     "accept compensation",
				ActorID:    in.FreelancerID,
				CreatedAt:  uc.Clock.Now(),
			})
			return AcceptContractOutput{}, wrapCompensationError(fmt.Errorf("sync proposal status: %w", err), revertErr, "revert contract activation")
		}
	}

	if updated.JobID > 0 && uc.Jobs != nil {
		err = uc.Jobs.SetInProgress(ctx, updated.JobID, updated.ClientID)
		if err != nil {
			revertErr := uc.Contracts.SetStatusForFreelancer(ctx, in.ContractID, in.FreelancerID, domain.StatusPendingAcceptance, uc.Clock.Now())
			if updated.ProposalID > 0 {
				revertProposalErr := uc.Proposals.MarkOfferSent(ctx, updated.ProposalID, updated.ClientID, "accept compensation")
				if revertProposalErr != nil {
					revertErr = fmt.Errorf("%w; proposal compensation: %v", revertErr, revertProposalErr)
				}
			}
			_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{
				ContractID: in.ContractID,
				Status:     domain.StatusPendingAcceptance,
				Reason:     "accept compensation",
				ActorID:    in.FreelancerID,
				CreatedAt:  uc.Clock.Now(),
			})
			return AcceptContractOutput{}, wrapCompensationError(fmt.Errorf("sync job status: %w", err), revertErr, "revert contract activation")
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
