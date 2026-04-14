package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/job/internal/domain"
)

type GetJobSummary struct {
	Jobs JobRepository
}

type GetJobSummaryInput struct {
	JobID int64
}

type GetJobSummaryOutput struct {
	Summary JobSummary
}

type JobSummary struct {
	JobID    int64
	ClientID string
	Status   string
	IsOpen   bool
	Found    bool
}

func (uc *GetJobSummary) Execute(ctx context.Context, in GetJobSummaryInput) (GetJobSummaryOutput, error) {
	if in.JobID <= 0 {
		return GetJobSummaryOutput{}, fmt.Errorf("job_id is required")
	}

	job, err := uc.Jobs.GetByID(ctx, in.JobID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return GetJobSummaryOutput{Summary: JobSummary{JobID: in.JobID, Found: false}}, nil
		}
		return GetJobSummaryOutput{}, err
	}

	return GetJobSummaryOutput{Summary: toJobSummary(job)}, nil
}

func toJobSummary(job domain.Job) JobSummary {
	return JobSummary{
		JobID:    job.ID,
		ClientID: job.ClientID.String(),
		Status:   job.Status,
		IsOpen:   strings.EqualFold(strings.TrimSpace(job.Status), domain.JobStatusOpen),
		Found:    true,
	}
}