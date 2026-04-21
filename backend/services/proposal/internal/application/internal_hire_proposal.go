package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type InternalHireProposal struct {
	Proposals ProposalRepository
	Clock     Clock
}

type InternalHireProposalInput struct {
	ProposalID int64
	ClientID   uuid.UUID
	RequestID  string
	Reason     string
}

type InternalHireProposalOutput struct {
	Proposal               domain.Proposal
	ReusedIdempotentResult bool
}

func (uc *InternalHireProposal) Execute(ctx context.Context, in InternalHireProposalInput) (InternalHireProposalOutput, error) {
	if in.ProposalID <= 0 {
		return InternalHireProposalOutput{}, fmt.Errorf("proposal_id is required")
	}
	if in.ClientID == uuid.Nil {
		return InternalHireProposalOutput{}, fmt.Errorf("client_id is required")
	}
	if strings.TrimSpace(in.RequestID) == "" {
		return InternalHireProposalOutput{}, fmt.Errorf("request_id is required")
	}
	if len(strings.TrimSpace(in.RequestID)) > 128 {
		return InternalHireProposalOutput{}, fmt.Errorf("request_id too long")
	}
	if len(strings.TrimSpace(in.Reason)) > 500 {
		return InternalHireProposalOutput{}, fmt.Errorf("reason too long")
	}

	current, err := uc.Proposals.GetByIDForClient(ctx, in.ProposalID, in.ClientID)
	if err != nil {
		return InternalHireProposalOutput{}, err
	}
	if !domain.CanTransition(current.Status, domain.StatusHired) {
		return InternalHireProposalOutput{}, fmt.Errorf("invalid proposal status transition")
	}

	proposal, reused, err := uc.Proposals.HireWithRequestID(ctx, in.ProposalID, in.ClientID, strings.TrimSpace(in.RequestID), strings.TrimSpace(in.Reason), uc.Clock.Now())
	if err != nil {
		return InternalHireProposalOutput{}, err
	}

	return InternalHireProposalOutput{Proposal: proposal, ReusedIdempotentResult: reused}, nil
}
