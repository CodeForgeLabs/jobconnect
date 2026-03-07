package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type SubmitProposal struct {
	Proposals ProposalRepository
	Jobs      JobReader
	Clock     Clock
}

type SubmitProposalInput struct {
	FreelancerID  uuid.UUID
	JobID         int64
	CoverLetter   string
	BidType       string
	BidAmount     float64
	EstimatedDays int32
	Attachments   []domain.Attachment
}

type SubmitProposalOutput struct {
	Proposal domain.Proposal
}

func (uc *SubmitProposal) Execute(ctx context.Context, in SubmitProposalInput) (SubmitProposalOutput, error) {
	if in.FreelancerID == uuid.Nil {
		return SubmitProposalOutput{}, fmt.Errorf("freelancer_id is required")
	}
	summary, err := uc.Jobs.GetJobSummary(ctx, in.JobID)
	if err != nil {
		return SubmitProposalOutput{}, err
	}
	if !summary.Found {
		return SubmitProposalOutput{}, fmt.Errorf("job not found")
	}
	if !summary.IsOpen || !strings.EqualFold(strings.TrimSpace(summary.Status), "open") {
		return SubmitProposalOutput{}, fmt.Errorf("job is not open")
	}

	hasActive, err := uc.Proposals.HasActiveProposal(ctx, in.JobID, in.FreelancerID)
	if err != nil {
		return SubmitProposalOutput{}, err
	}
	if hasActive {
		return SubmitProposalOutput{}, fmt.Errorf("active proposal already exists for this job")
	}

	now := uc.Clock.Now()
	p := domain.Proposal{
		JobID:         in.JobID,
		ClientID:      summary.ClientID,
		FreelancerID:  in.FreelancerID,
		CoverLetter:   strings.TrimSpace(in.CoverLetter),
		BidType:       strings.ToLower(strings.TrimSpace(in.BidType)),
		BidAmount:     in.BidAmount,
		EstimatedDays: in.EstimatedDays,
		Attachments:   in.Attachments,
		Status:        domain.StatusSent,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := domain.ValidateForSubmit(p); err != nil {
		return SubmitProposalOutput{}, err
	}

	id, err := uc.Proposals.Create(ctx, p)
	if err != nil {
		return SubmitProposalOutput{}, err
	}

	persisted, err := uc.Proposals.GetByID(ctx, id)
	if err != nil {
		return SubmitProposalOutput{}, err
	}

	return SubmitProposalOutput{Proposal: persisted}, nil
}
