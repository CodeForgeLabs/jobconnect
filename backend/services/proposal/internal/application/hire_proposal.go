package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type HireProposal struct {
	Proposals ProposalRepository
	Jobs      JobReader
	Clock     Clock
}

type HireProposalInput struct {
	ProposalID int64
	ClientID   uuid.UUID
	RequestID  string
	Reason     string
}

type HireProposalOutput struct {
	Proposal              domain.Proposal
	ReusedIdempotentResult bool
}

func (uc *HireProposal) Execute(ctx context.Context, in HireProposalInput) (HireProposalOutput, error) {
	if in.ProposalID <= 0 {
		return HireProposalOutput{}, fmt.Errorf("proposal_id is required")
	}
	if in.ClientID == uuid.Nil {
		return HireProposalOutput{}, fmt.Errorf("client_id is required")
	}
	if strings.TrimSpace(in.RequestID) == "" {
		return HireProposalOutput{}, fmt.Errorf("request_id is required")
	}
	if len(strings.TrimSpace(in.RequestID)) > 128 {
		return HireProposalOutput{}, fmt.Errorf("request_id too long")
	}
	if len(strings.TrimSpace(in.Reason)) > 500 {
		return HireProposalOutput{}, fmt.Errorf("reason too long")
	}

	current, err := uc.Proposals.GetByIDForClient(ctx, in.ProposalID, in.ClientID)
	if err != nil {
		return HireProposalOutput{}, err
	}

	if !domain.CanTransition(current.Status, domain.StatusHired) {
		return HireProposalOutput{}, fmt.Errorf("invalid proposal status transition")
	}

	hasHired, err := uc.Proposals.HasHiredProposalForJob(ctx, current.JobID)
	if err != nil {
		return HireProposalOutput{}, err
	}
	if hasHired {
		return HireProposalOutput{}, fmt.Errorf("job already has a hired proposal")
	}

	summary, err := uc.Jobs.GetJobSummary(ctx, current.JobID)
	if err != nil {
		return HireProposalOutput{}, err
	}
	if !summary.Found {
		return HireProposalOutput{}, fmt.Errorf("job not found")
	}
	if summary.ClientID != in.ClientID {
		return HireProposalOutput{}, fmt.Errorf("proposal does not belong to this client")
	}
	if !summary.IsOpen {
		return HireProposalOutput{}, fmt.Errorf("job is not open")
	}

	proposal, reused, err := uc.Proposals.HireWithRequestID(ctx, in.ProposalID, in.ClientID, strings.TrimSpace(in.RequestID), strings.TrimSpace(in.Reason), uc.Clock.Now())
	if err != nil {
		return HireProposalOutput{}, err
	}
	return HireProposalOutput{Proposal: proposal, ReusedIdempotentResult: reused}, nil
}
