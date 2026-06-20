package application

import (
	"context"
	"fmt"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type PauseJob struct {
	Jobs  JobRepository
	Clock Clock
}

type PauseJobInput struct {
	JobID    int64
	ClientID uuid.UUID
}

type PauseJobOutput struct {
	Job domain.Job
}

func (uc *PauseJob) Execute(ctx context.Context, in PauseJobInput) (PauseJobOutput, error) {
	if in.JobID <= 0 {
		return PauseJobOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return PauseJobOutput{}, fmt.Errorf("client_id is required")
	}
	job, err := uc.Jobs.Pause(ctx, in.JobID, in.ClientID, uc.Clock.Now())
	if err != nil {
		return PauseJobOutput{}, err
	}
	return PauseJobOutput{Job: job}, nil
}

type ReopenJob struct {
	Jobs  JobRepository
	Clock Clock
}

type ReopenJobInput struct {
	JobID    int64
	ClientID uuid.UUID
}

type ReopenJobOutput struct {
	Job domain.Job
}

func (uc *ReopenJob) Execute(ctx context.Context, in ReopenJobInput) (ReopenJobOutput, error) {
	if in.JobID <= 0 {
		return ReopenJobOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return ReopenJobOutput{}, fmt.Errorf("client_id is required")
	}
	job, err := uc.Jobs.Reopen(ctx, in.JobID, in.ClientID, uc.Clock.Now())
	if err != nil {
		return ReopenJobOutput{}, err
	}
	return ReopenJobOutput{Job: job}, nil
}

type MarkJobFilled struct {
	Jobs  JobRepository
	Clock Clock
}

type MarkJobFilledInput struct {
	JobID    int64
	ClientID uuid.UUID
}

type MarkJobFilledOutput struct {
	Job domain.Job
}

func (uc *MarkJobFilled) Execute(ctx context.Context, in MarkJobFilledInput) (MarkJobFilledOutput, error) {
	if in.JobID <= 0 {
		return MarkJobFilledOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return MarkJobFilledOutput{}, fmt.Errorf("client_id is required")
	}
	job, err := uc.Jobs.MarkFilled(ctx, in.JobID, in.ClientID, uc.Clock.Now())
	if err != nil {
		return MarkJobFilledOutput{}, err
	}
	return MarkJobFilledOutput{Job: job}, nil
}
