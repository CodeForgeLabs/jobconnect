package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type ModifyProposal struct {
	Proposals ProposalRepository
	Clock     Clock
}

type ModifyProposalInput struct {
	ProposalID    int64
	FreelancerID  uuid.UUID
	CoverLetter   string
	BidAmount     float64
	EstimatedDays int32
	Attachments   []domain.Attachment
}

type ModifyProposalOutput struct {
	Proposal domain.Proposal
}

func (uc *ModifyProposal) Execute(ctx context.Context, in ModifyProposalInput) (ModifyProposalOutput, error) {
	if in.ProposalID <= 0 {
		return ModifyProposalOutput{}, fmt.Errorf("proposal_id is required")
	}
	if in.FreelancerID == uuid.Nil {
		return ModifyProposalOutput{}, fmt.Errorf("freelancer_id is required")
	}
	if err := domain.ValidateForModify(in.CoverLetter, in.BidAmount, in.EstimatedDays, in.Attachments); err != nil {
		return ModifyProposalOutput{}, err
	}

	current, err := uc.Proposals.GetByIDForFreelancer(ctx, in.ProposalID, in.FreelancerID)
	if err != nil {
		return ModifyProposalOutput{}, err
	}
	if !domain.CanFreelancerModify(current.Status) {
		return ModifyProposalOutput{}, fmt.Errorf("proposal cannot be modified in current status")
	}

	err = uc.Proposals.UpdateEditable(
		ctx,
		in.ProposalID,
		in.FreelancerID,
		strings.TrimSpace(in.CoverLetter),
		in.BidAmount,
		in.EstimatedDays,
		in.Attachments,
		uc.Clock.Now(),
	)
	if err != nil {
		return ModifyProposalOutput{}, err
	}

	persisted, err := uc.Proposals.GetByIDForFreelancer(ctx, in.ProposalID, in.FreelancerID)
	if err != nil {
		return ModifyProposalOutput{}, err
	}
	return ModifyProposalOutput{Proposal: persisted}, nil
}
