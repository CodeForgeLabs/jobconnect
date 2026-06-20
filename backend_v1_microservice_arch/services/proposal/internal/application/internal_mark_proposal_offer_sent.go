package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type InternalMarkProposalOfferSent struct {
	Proposals ProposalRepository
	Clock     Clock
}

type InternalMarkProposalOfferSentInput struct {
	ProposalID int64
	ClientID   uuid.UUID
	Reason     string
}

type InternalMarkProposalOfferSentOutput struct {
	Proposal domain.Proposal
}

func (uc *InternalMarkProposalOfferSent) Execute(ctx context.Context, in InternalMarkProposalOfferSentInput) (InternalMarkProposalOfferSentOutput, error) {
	if in.ProposalID <= 0 {
		return InternalMarkProposalOfferSentOutput{}, fmt.Errorf("proposal_id is required")
	}
	if in.ClientID == uuid.Nil {
		return InternalMarkProposalOfferSentOutput{}, fmt.Errorf("client_id is required")
	}
	if len(strings.TrimSpace(in.Reason)) > 500 {
		return InternalMarkProposalOfferSentOutput{}, fmt.Errorf("reason too long")
	}

	proposal, err := uc.Proposals.MarkOfferSent(ctx, in.ProposalID, in.ClientID, strings.TrimSpace(in.Reason), uc.Clock.Now())
	if err != nil {
		return InternalMarkProposalOfferSentOutput{}, err
	}
	return InternalMarkProposalOfferSentOutput{Proposal: proposal}, nil
}
