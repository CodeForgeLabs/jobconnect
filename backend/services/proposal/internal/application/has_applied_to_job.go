package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type HasAppliedToJob struct {
	Proposals ProposalRepository
}

type HasAppliedToJobInput struct {
	FreelancerID uuid.UUID
	JobID        int64
}

type HasAppliedToJobOutput struct {
	HasApplied bool
	ProposalID int64
	Status     string
}

func (uc *HasAppliedToJob) Execute(ctx context.Context, in HasAppliedToJobInput) (HasAppliedToJobOutput, error) {
	if in.FreelancerID == uuid.Nil {
		return HasAppliedToJobOutput{}, fmt.Errorf("freelancer_id is required")
	}
	if in.JobID <= 0 {
		return HasAppliedToJobOutput{}, fmt.Errorf("job_id is required")
	}

	p, err := uc.Proposals.GetLatestByJobForFreelancer(ctx, in.JobID, in.FreelancerID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return HasAppliedToJobOutput{HasApplied: false}, nil
		}
		return HasAppliedToJobOutput{}, err
	}

	return HasAppliedToJobOutput{
		HasApplied: true,
		ProposalID: p.ID,
		Status:     p.Status,
	}, nil
}
