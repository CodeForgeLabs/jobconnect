package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type UpdateJob struct {
	Jobs  JobRepository
	Clock Clock
}

type UpdateJobInput struct {
	JobID       int64
	ClientID    uuid.UUID
	Title       *string
	Description *string
	RequiredSkills []string
	JobType     *string
	BudgetFixed *float64
	HourlyRate  *float64
	Currency    *string
	Deadline    *int64
	Attachments []domain.Attachment
}

type UpdateJobOutput struct {
	Job domain.Job
}

func (uc *UpdateJob) Execute(ctx context.Context, in UpdateJobInput) (UpdateJobOutput, error) {
	if in.JobID <= 0 {
		return UpdateJobOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return UpdateJobOutput{}, fmt.Errorf("client_id is required")
	}

	// Fetch current job (must belong to the client).
	job, err := uc.Jobs.GetByIDForClient(ctx, in.JobID, in.ClientID)
	if err != nil {
		return UpdateJobOutput{}, err
	}

	// Only open jobs can be updated.
	if job.Status != domain.JobStatusOpen {
		return UpdateJobOutput{}, fmt.Errorf("only open jobs can be updated")
	}

	// Apply partial updates.
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return UpdateJobOutput{}, fmt.Errorf("title is required")
		}
		if len(title) > 160 {
			return UpdateJobOutput{}, fmt.Errorf("title too long")
		}
		job.Title = title
	}
	if in.Description != nil {
		desc := strings.TrimSpace(*in.Description)
		if desc == "" {
			return UpdateJobOutput{}, fmt.Errorf("description is required")
		}
		if len(desc) > 10000 {
			return UpdateJobOutput{}, fmt.Errorf("description too long")
		}
		job.Description = desc
	}
	if len(in.RequiredSkills) > 0 {
		if len(in.RequiredSkills) > 100 {
			return UpdateJobOutput{}, fmt.Errorf("too many required_skills")
		}
		for _, s := range in.RequiredSkills {
			if strings.TrimSpace(s) == "" {
				return UpdateJobOutput{}, fmt.Errorf("required_skills contains empty value")
			}
		}
		job.RequiredSkills = in.RequiredSkills
	}
	if in.JobType != nil {
		jt := strings.ToLower(strings.TrimSpace(*in.JobType))
		if err := domain.ValidateJobType(jt); err != nil {
			return UpdateJobOutput{}, err
		}
		job.JobType = jt
	}
	if in.BudgetFixed != nil {
		job.BudgetFixed = *in.BudgetFixed
	}
	if in.HourlyRate != nil {
		job.HourlyRate = *in.HourlyRate
	}
	if in.Currency != nil {
		cur := strings.ToUpper(strings.TrimSpace(*in.Currency))
		if cur == "" {
			return UpdateJobOutput{}, fmt.Errorf("currency is required")
		}
		if len(cur) > 8 {
			return UpdateJobOutput{}, fmt.Errorf("currency too long")
		}
		job.Currency = cur
	}
	if in.Deadline != nil && *in.Deadline > 0 {
		deadline := time.Unix(*in.Deadline, 0).UTC()
		if !deadline.After(uc.Clock.Now()) {
			return UpdateJobOutput{}, fmt.Errorf("deadline must be in the future")
		}
		job.Deadline = &deadline
	}
	if len(in.Attachments) > 0 {
		if len(in.Attachments) > 20 {
			return UpdateJobOutput{}, fmt.Errorf("too many attachments")
		}
		for _, a := range in.Attachments {
			if strings.TrimSpace(a.FileName) == "" {
				return UpdateJobOutput{}, fmt.Errorf("attachment file_name is required")
			}
			if strings.TrimSpace(a.ContentType) == "" {
				return UpdateJobOutput{}, fmt.Errorf("attachment content_type is required")
			}
			if strings.TrimSpace(a.URL) == "" {
				return UpdateJobOutput{}, fmt.Errorf("attachment url is required")
			}
		}
	}

	// Re-validate budget/type consistency after partial updates.
	if job.JobType == domain.JobTypeFixed {
		if job.BudgetFixed <= 0 {
			return UpdateJobOutput{}, fmt.Errorf("budget_fixed must be greater than zero")
		}
		if job.HourlyRate > 0 {
			return UpdateJobOutput{}, fmt.Errorf("hourly_rate must be empty for fixed jobs")
		}
	}
	if job.JobType == domain.JobTypeHourly {
		if job.HourlyRate <= 0 {
			return UpdateJobOutput{}, fmt.Errorf("hourly_rate must be greater than zero")
		}
		if job.BudgetFixed > 0 {
			return UpdateJobOutput{}, fmt.Errorf("budget_fixed must be empty for hourly jobs")
		}
	}

	job.UpdatedAt = uc.Clock.Now()

	updated, err := uc.Jobs.Update(ctx, job)
	if err != nil {
		return UpdateJobOutput{}, fmt.Errorf("update job: %w", err)
	}

	return UpdateJobOutput{Job: updated}, nil
}
