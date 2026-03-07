package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type GetProposal struct {
	Proposals ProposalRepository
}

type GetProposalInput struct {
	ProposalID int64
	ActorID    uuid.UUID
	ActorRole  string // client | freelancer
}

type GetProposalOutput struct {
	Proposal domain.Proposal
}

func (uc *GetProposal) Execute(ctx context.Context, in GetProposalInput) (GetProposalOutput, error) {
	if in.ProposalID <= 0 {
		return GetProposalOutput{}, fmt.Errorf("proposal_id is required")
	}
	if in.ActorID == uuid.Nil {
		return GetProposalOutput{}, fmt.Errorf("actor_id is required")
	}

	role := strings.ToLower(strings.TrimSpace(in.ActorRole))
	switch role {
	case "client":
		p, err := uc.Proposals.GetByIDForClient(ctx, in.ProposalID, in.ActorID)
		if err != nil {
			return GetProposalOutput{}, err
		}
		return GetProposalOutput{Proposal: p}, nil
	case "freelancer":
		p, err := uc.Proposals.GetByIDForFreelancer(ctx, in.ProposalID, in.ActorID)
		if err != nil {
			return GetProposalOutput{}, err
		}
		return GetProposalOutput{Proposal: p}, nil
	default:
		return GetProposalOutput{}, fmt.Errorf("unsupported actor role")
	}
}
