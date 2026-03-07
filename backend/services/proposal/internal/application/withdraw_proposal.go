package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type WithdrawProposal struct {
	Proposals ProposalRepository
	Clock     Clock
}

type WithdrawProposalInput struct {
	ProposalID   int64
	FreelancerID uuid.UUID
	Reason       string
}

type WithdrawProposalOutput struct {
	Withdrawn bool
}

func (uc *WithdrawProposal) Execute(ctx context.Context, in WithdrawProposalInput) (WithdrawProposalOutput, error) {
	if in.ProposalID <= 0 {
		return WithdrawProposalOutput{}, fmt.Errorf("proposal_id is required")
	}
	if in.FreelancerID == uuid.Nil {
		return WithdrawProposalOutput{}, fmt.Errorf("freelancer_id is required")
	}
	if len(strings.TrimSpace(in.Reason)) > 500 {
		return WithdrawProposalOutput{}, fmt.Errorf("reason too long")
	}

	current, err := uc.Proposals.GetByIDForFreelancer(ctx, in.ProposalID, in.FreelancerID)
	if err != nil {
		return WithdrawProposalOutput{}, err
	}
	if !domain.CanFreelancerWithdraw(current.Status) {
		return WithdrawProposalOutput{}, fmt.Errorf("proposal cannot be withdrawn in current status")
	}

	if err := uc.Proposals.Withdraw(ctx, in.ProposalID, in.FreelancerID, strings.TrimSpace(in.Reason), uc.Clock.Now()); err != nil {
		return WithdrawProposalOutput{}, err
	}

	return WithdrawProposalOutput{Withdrawn: true}, nil
}
