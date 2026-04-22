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
	if uc.Contracts == nil || uc.Proposals == nil || uc.Actors == nil || uc.Clock == nil {
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
	if !strings.EqualFold(proposal.Status, "hired") {
		return CreateContractOutput{}, fmt.Errorf("proposal must be reserved before sending an offer")
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
		case domain.StatusPendingAcceptance, domain.StatusDeclined, domain.StatusRevoked:
			c.ID = existing.ID
			if err := uc.Contracts.UpdateOfferForClient(ctx, c); err != nil {
				return CreateContractOutput{}, err
			}
			_ = uc.Contracts.AppendStatusHistory(ctx, domain.StatusHistoryEntry{
				ContractID: existing.ID,
				Status:     domain.StatusPendingAcceptance,
				Reason:     "offer sent",
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

	id, err := uc.Contracts.Create(ctx, c)
	if err != nil {
		return CreateContractOutput{}, err
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
