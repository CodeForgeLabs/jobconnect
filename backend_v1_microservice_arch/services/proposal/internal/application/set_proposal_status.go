package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type SetProposalStatus struct {
	Proposals ProposalRepository
	Clock     Clock
}

type SetProposalStatusInput struct {
	ProposalID int64
	ClientID   uuid.UUID
	Status     string // shortlisted | rejected | hired
	Reason     string
}

type SetProposalStatusOutput struct {
	Proposal domain.Proposal
}

func (uc *SetProposalStatus) Execute(ctx context.Context, in SetProposalStatusInput) (SetProposalStatusOutput, error) {
	if in.ProposalID <= 0 {
		return SetProposalStatusOutput{}, fmt.Errorf("proposal_id is required")
	}
	if in.ClientID == uuid.Nil {
		return SetProposalStatusOutput{}, fmt.Errorf("client_id is required")
	}
	if len(strings.TrimSpace(in.Reason)) > 500 {
		return SetProposalStatusOutput{}, fmt.Errorf("reason too long")
	}

	nextStatus := strings.ToLower(strings.TrimSpace(in.Status))
	if nextStatus == domain.StatusWithdrawn || nextStatus == domain.StatusSent || nextStatus == domain.StatusOfferSent || nextStatus == domain.StatusHired {
		return SetProposalStatusOutput{}, fmt.Errorf("invalid target status")
	}
	if err := domain.ValidateStatus(nextStatus); err != nil {
		return SetProposalStatusOutput{}, err
	}

	current, err := uc.Proposals.GetByIDForClient(ctx, in.ProposalID, in.ClientID)
	if err != nil {
		return SetProposalStatusOutput{}, err
	}
	if !domain.CanTransition(current.Status, nextStatus) {
		return SetProposalStatusOutput{}, fmt.Errorf("invalid proposal status transition")
	}

	if err := uc.Proposals.SetStatus(ctx, in.ProposalID, in.ClientID, nextStatus, strings.TrimSpace(in.Reason), uc.Clock.Now()); err != nil {
		return SetProposalStatusOutput{}, err
	}

	persisted, err := uc.Proposals.GetByIDForClient(ctx, in.ProposalID, in.ClientID)
	if err != nil {
		return SetProposalStatusOutput{}, err
	}
	return SetProposalStatusOutput{Proposal: persisted}, nil
}
