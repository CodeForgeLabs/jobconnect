package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type CreateJob struct {
	Jobs  JobRepository
	Clock Clock
}

type CreateJobInput struct {
	ClientID       uuid.UUID
	Title          string
	Description    string
	RequiredSkills []string
	JobType        string
	BudgetFixed    float64
	HourlyRate     float64
	Deadline       *int64
	Attachments    []domain.Attachment
}

type CreateJobOutput struct {
	Job domain.Job
}

func (uc *CreateJob) Execute(ctx context.Context, in CreateJobInput) (CreateJobOutput, error) {
	now := uc.Clock.Now()
	job := domain.Job{
		ClientID:       in.ClientID,
		Title:          strings.TrimSpace(in.Title),
		Description:    strings.TrimSpace(in.Description),
		RequiredSkills: in.RequiredSkills,
		JobType:        strings.ToLower(strings.TrimSpace(in.JobType)),
		BudgetFixed:    in.BudgetFixed,
		HourlyRate:     in.HourlyRate,
		Visibility:     domain.VisibilityPublic,
		Attachments:    in.Attachments,
		Status:         domain.JobStatusOpen,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Set budget range based on job type.
	if job.JobType == domain.JobTypeFixed {
		job.BudgetMin = in.BudgetFixed
		job.BudgetMax = in.BudgetFixed
	} else if job.JobType == domain.JobTypeHourly {
		job.BudgetMin = in.HourlyRate
		job.BudgetMax = in.HourlyRate
	}

	if in.Deadline != nil && *in.Deadline > 0 {
		deadline := time.Unix(*in.Deadline, 0).UTC()
		job.Deadline = &deadline
	}

	if err := domain.ValidateCreate(job, now); err != nil {
		return CreateJobOutput{}, err
	}

	id, err := uc.Jobs.Create(ctx, job)
	if err != nil {
		return CreateJobOutput{}, err
	}

	persisted, err := uc.Jobs.GetByID(ctx, id)
	if err != nil {
		return CreateJobOutput{}, fmt.Errorf("create job: %w", err)
	}

	return CreateJobOutput{Job: persisted}, nil
}
