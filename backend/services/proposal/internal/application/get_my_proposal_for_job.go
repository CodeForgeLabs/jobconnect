package application

import (
	"context"
	"fmt"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type GetMyProposalForJob struct {
	Proposals ProposalRepository
}

type GetMyProposalForJobInput struct {
	FreelancerID uuid.UUID
	JobID        int64
}

type GetMyProposalForJobOutput struct {
	Proposal domain.Proposal
}

func (uc *GetMyProposalForJob) Execute(ctx context.Context, in GetMyProposalForJobInput) (GetMyProposalForJobOutput, error) {
	if in.FreelancerID == uuid.Nil {
		return GetMyProposalForJobOutput{}, fmt.Errorf("freelancer_id is required")
	}
	if in.JobID <= 0 {
		return GetMyProposalForJobOutput{}, fmt.Errorf("job_id is required")
	}

	p, err := uc.Proposals.GetLatestByJobForFreelancer(ctx, in.JobID, in.FreelancerID)
	if err != nil {
		return GetMyProposalForJobOutput{}, err
	}
	return GetMyProposalForJobOutput{Proposal: p}, nil
}
