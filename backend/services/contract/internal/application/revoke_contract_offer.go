package application

import (
	"context"
	"fmt"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type RevokeContractOffer struct {
	Contracts ContractRepository
	Proposals ProposalSync
	Actors    ActorPolicy
	Clock     Clock
}

type RevokeContractOfferInput struct {
	ContractID int64
	ClientID   uuid.UUID
	Reason     string
}

type RevokeContractOfferOutput struct {
	Contract domain.Contract
}

func (uc *RevokeContractOffer) Execute(ctx context.Context, in RevokeContractOfferInput) (RevokeContractOfferOutput, error) {
	if uc.Contracts == nil || uc.Proposals == nil || uc.Actors == nil || uc.Clock == nil {
		return RevokeContractOfferOutput{}, fmt.Errorf("revoke contract dependencies are not configured")
	}
	if in.ContractID <= 0 {
		return RevokeContractOfferOutput{}, fmt.Errorf("contract_id is required")
	}
	if in.ClientID == uuid.Nil {
		return RevokeContractOfferOutput{}, fmt.Errorf("client_id is required")
	}
	if err := uc.Actors.EnsureClientCanHire(ctx, in.ClientID); err != nil {
		return RevokeContractOfferOutput{}, err
	}

	current, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ClientID)
	if err != nil {
		return RevokeContractOfferOutput{}, err
	}
	if !domain.CanRevoke(current.Status) {
		return RevokeContractOfferOutput{}, fmt.Errorf("contract is not in a revoke-able state")
	}
	now := uc.Clock.Now()
	if err := uc.Contracts.SetStatusForClient(ctx, in.ContractID, in.ClientID, domain.StatusRevoked, now); err != nil {
		return RevokeContractOfferOutput{}, err
	}
	if current.ProposalID > 0 {
		if err := uc.Proposals.ReleaseOffer(ctx, current.ProposalID, current.ClientID, in.Reason); err != nil {
			revertErr := uc.Contracts.SetStatusForClient(ctx, in.ContractID, in.ClientID, domain.StatusPendingAcceptance, uc.Clock.Now())
			_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{
				ContractID: in.ContractID,
				Status:     domain.StatusPendingAcceptance,
				Reason:     "revoke compensation",
				ActorID:    in.ClientID,
				CreatedAt:  uc.Clock.Now(),
			})
			return RevokeContractOfferOutput{}, wrapCompensationError(fmt.Errorf("sync proposal status: %w", err), revertErr, "revert contract revocation")
		}
	}
	_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{
		ContractID: in.ContractID,
		Status:     domain.StatusRevoked,
		Reason:     in.Reason,
		ActorID:    in.ClientID,
		CreatedAt:  now,
	})
	updated, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.ClientID)
	if err != nil {
		return RevokeContractOfferOutput{}, err
	}
	return RevokeContractOfferOutput{Contract: updated}, nil
}

type GetJobOfferState struct {
	Contracts ContractRepository
}

type GetJobOfferStateInput struct {
	JobID     int64
	ClientID  uuid.UUID
	ActorRole string
}

type GetJobOfferStateOutput struct {
	State domain.JobOfferState
}

func (uc *GetJobOfferState) Execute(ctx context.Context, in GetJobOfferStateInput) (GetJobOfferStateOutput, error) {
	if uc.Contracts == nil {
		return GetJobOfferStateOutput{}, fmt.Errorf("get job offer state dependencies are not configured")
	}
	if in.JobID <= 0 {
		return GetJobOfferStateOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return GetJobOfferStateOutput{}, fmt.Errorf("client_id is required")
	}
	state, err := uc.Contracts.GetJobOfferState(ctx, in.JobID, in.ClientID)
	if err != nil {
		return GetJobOfferStateOutput{}, err
	}
	return GetJobOfferStateOutput{State: state}, nil
}
