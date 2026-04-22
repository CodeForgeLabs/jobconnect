package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type ReleaseHiredProposal struct {
	Proposals ProposalRepository
	Clock     Clock
}

type ReleaseHiredProposalInput struct {
	ProposalID int64
	ClientID   uuid.UUID
	Reason     string
}

type ReleaseHiredProposalOutput struct {
	Proposal domain.Proposal
}

func (uc *ReleaseHiredProposal) Execute(ctx context.Context, in ReleaseHiredProposalInput) (ReleaseHiredProposalOutput, error) {
	if in.ProposalID <= 0 {
		return ReleaseHiredProposalOutput{}, fmt.Errorf("proposal_id is required")
	}
	if in.ClientID == uuid.Nil {
		return ReleaseHiredProposalOutput{}, fmt.Errorf("client_id is required")
	}
	if len(strings.TrimSpace(in.Reason)) > 500 {
		return ReleaseHiredProposalOutput{}, fmt.Errorf("reason too long")
	}

	current, err := uc.Proposals.GetByIDForClient(ctx, in.ProposalID, in.ClientID)
	if err != nil {
		return ReleaseHiredProposalOutput{}, err
	}
	switch current.Status {
	case domain.StatusOfferSent:
		if err := uc.Proposals.SetStatus(ctx, in.ProposalID, in.ClientID, domain.StatusShortlisted, strings.TrimSpace(in.Reason), uc.Clock.Now()); err != nil {
			return ReleaseHiredProposalOutput{}, err
		}
	case domain.StatusHired:
		if err := uc.Proposals.RevertHire(ctx, in.ProposalID, in.ClientID, strings.TrimSpace(in.Reason), uc.Clock.Now()); err != nil {
			return ReleaseHiredProposalOutput{}, err
		}
	case domain.StatusShortlisted:
		// Treat repeated release calls as idempotent once the proposal is unlocked.
	default:
		return ReleaseHiredProposalOutput{}, fmt.Errorf("proposal cannot be released in current status")
	}

	updated, err := uc.Proposals.GetByIDForClient(ctx, in.ProposalID, in.ClientID)
	if err != nil {
		return ReleaseHiredProposalOutput{}, err
	}
	return ReleaseHiredProposalOutput{Proposal: updated}, nil
}
