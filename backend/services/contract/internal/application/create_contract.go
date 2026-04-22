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
	Proposals ProposalSync
	Jobs      JobReader
	Actors    ActorPolicy
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
	HourlyRate      float64
	FixedTotal      float64
	WeeklyHourLimit int32
	Milestones      []domain.Milestone
}

type CreateContractOutput struct {
	Contract domain.Contract
}

func (uc *CreateContract) Execute(ctx context.Context, in CreateContractInput) (CreateContractOutput, error) {
	if uc.Contracts == nil || uc.Proposals == nil || uc.Jobs == nil || uc.Actors == nil || uc.Clock == nil {
		return CreateContractOutput{}, fmt.Errorf("create contract dependencies are not configured")
	}
	if err := uc.Actors.EnsureClientCanHire(ctx, in.ClientID); err != nil {
		return CreateContractOutput{}, err
	}
	if err := uc.Actors.EnsureFreelancerCanWork(ctx, in.FreelancerID); err != nil {
		return CreateContractOutput{}, err
	}

	proposal, err := uc.Proposals.GetProposal(ctx, in.ProposalID, in.ClientID)
	if err != nil {
		return CreateContractOutput{}, err
	}
	if proposal.JobID != in.JobID {
		return CreateContractOutput{}, fmt.Errorf("proposal does not belong to job")
	}
	if proposal.FreelancerID != in.FreelancerID.String() {
		return CreateContractOutput{}, fmt.Errorf("proposal does not belong to freelancer")
	}
	if !strings.EqualFold(proposal.Status, "sent") && !strings.EqualFold(proposal.Status, "shortlisted") && !strings.EqualFold(proposal.Status, "hired") {
		return CreateContractOutput{}, fmt.Errorf("proposal is not eligible for offer")
	}

	job, err := uc.Jobs.GetSummary(ctx, in.JobID, in.ClientID)
	if err != nil {
		return CreateContractOutput{}, err
	}
	if !job.Found || !strings.EqualFold(job.ClientID, in.ClientID.String()) {
		return CreateContractOutput{}, fmt.Errorf("job not found")
	}
	if !job.IsOpen {
		return CreateContractOutput{}, fmt.Errorf("job must be open to send an offer")
	}
	offerState, err := uc.Contracts.GetJobOfferState(ctx, in.JobID, in.ClientID)
	if err != nil {
		return CreateContractOutput{}, err
	}
	if offerState.HasActiveContract {
		return CreateContractOutput{}, fmt.Errorf("job already has an active contract")
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

	existing, err := uc.Contracts.GetByProposalID(ctx, in.ProposalID)
	switch {
	case err == nil:
		switch existing.Status {
		case domain.StatusPendingAcceptance:
			return CreateContractOutput{}, fmt.Errorf("job already has a pending offer")
		case domain.StatusDeclined, domain.StatusRevoked:
			if offerState.HasPendingOffer && offerState.PendingContractID != existing.ID {
				return CreateContractOutput{}, fmt.Errorf("job already has a pending offer")
			}
			c.ID = existing.ID
			if !strings.EqualFold(proposal.Status, "hired") {
				if err := uc.Proposals.SetHired(ctx, in.ProposalID, in.ClientID, "offer sent"); err != nil {
					return CreateContractOutput{}, fmt.Errorf("sync proposal status: %w", err)
				}
			}
			if err := uc.Contracts.UpdateOfferForClient(ctx, c); err != nil {
				if !strings.EqualFold(proposal.Status, "hired") {
					revertErr := uc.Proposals.ReleaseHired(ctx, in.ProposalID, in.ClientID, "offer resend compensation")
					return CreateContractOutput{}, wrapCompensationError(fmt.Errorf("update resent offer: %w", err), revertErr, "revert proposal hire after offer resend failure")
				}
				return CreateContractOutput{}, err
			}
			_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{
				ContractID: existing.ID,
				Status:     domain.StatusPendingAcceptance,
				Reason:     "offer resent",
				ActorID:    in.ClientID,
				CreatedAt:  now,
			})
			persisted, getErr := uc.Contracts.GetByID(ctx, existing.ID)
			if getErr != nil {
				return CreateContractOutput{}, getErr
			}
			return CreateContractOutput{Contract: persisted}, nil
		default:
			return CreateContractOutput{}, fmt.Errorf("contract already exists for proposal")
		}
	case !strings.Contains(strings.ToLower(err.Error()), "not found"):
		return CreateContractOutput{}, err
	}
	if offerState.HasPendingOffer {
		return CreateContractOutput{}, fmt.Errorf("job already has a pending offer")
	}

	id, err := uc.Contracts.Create(ctx, c)
	if err != nil {
		return CreateContractOutput{}, err
	}
	if !strings.EqualFold(proposal.Status, "hired") {
		if err := uc.Proposals.SetHired(ctx, in.ProposalID, in.ClientID, "offer sent"); err != nil {
			revertErr := uc.Contracts.SetStatusForClient(ctx, id, in.ClientID, domain.StatusRevoked, now)
			_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{
				ContractID: id,
				Status:     domain.StatusRevoked,
				Reason:     "offer send compensation",
				ActorID:    in.ClientID,
				CreatedAt:  now,
			})
			return CreateContractOutput{}, wrapCompensationError(err, revertErr, "revoke offer after proposal sync failure")
		}
	}
	_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{
		ContractID: id,
		Status:     domain.StatusPendingAcceptance,
		Reason:     "offer sent",
		ActorID:    in.ClientID,
		CreatedAt:  now,
	})
	persisted, err := uc.Contracts.GetByID(ctx, id)
	if err != nil {
		return CreateContractOutput{}, err
	}
	return CreateContractOutput{Contract: persisted}, nil
}
