package application

import (
	"context"
	"fmt"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type GetJob struct {
	Jobs JobRepository
}

type GetJobInput struct {
	JobID    int64
	ClientID uuid.UUID
}

type GetJobOutput struct {
	Job domain.Job
}

func (uc *GetJob) Execute(ctx context.Context, in GetJobInput) (GetJobOutput, error) {
	if in.JobID <= 0 {
		return GetJobOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return GetJobOutput{}, fmt.Errorf("client_id is required")
	}
	job, err := uc.Jobs.GetByIDForClient(ctx, in.JobID, in.ClientID)
	if err != nil {
		return GetJobOutput{}, err
	}
	return GetJobOutput{Job: job}, nil
}
