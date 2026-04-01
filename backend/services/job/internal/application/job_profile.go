package application

import (
	"context"
	"fmt"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type SetJobVisibility struct {
	Jobs  JobRepository
	Clock Clock
}

type SetJobVisibilityInput struct {
	JobID      int64
	ClientID   uuid.UUID
	Visibility string
}

type SetJobVisibilityOutput struct {
	Job domain.Job
}

func (uc *SetJobVisibility) Execute(ctx context.Context, in SetJobVisibilityInput) (SetJobVisibilityOutput, error) {
	if in.JobID <= 0 {
		return SetJobVisibilityOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return SetJobVisibilityOutput{}, fmt.Errorf("client_id is required")
	}
	visibility, err := domain.ValidateVisibility(in.Visibility)
	if err != nil {
		return SetJobVisibilityOutput{}, err
	}
	if visibility == "" {
		return SetJobVisibilityOutput{}, fmt.Errorf("visibility is required")
	}

	job, err := uc.Jobs.SetVisibility(ctx, in.JobID, in.ClientID, visibility, uc.Clock.Now())
	if err != nil {
		return SetJobVisibilityOutput{}, err
	}
	return SetJobVisibilityOutput{Job: job}, nil
}

type SetJobBudgetRange struct {
	Jobs  JobRepository
	Clock Clock
}

type SetJobBudgetRangeInput struct {
	JobID     int64
	ClientID  uuid.UUID
	BudgetMin float64
	BudgetMax float64
}

type SetJobBudgetRangeOutput struct {
	Job domain.Job
}

func (uc *SetJobBudgetRange) Execute(ctx context.Context, in SetJobBudgetRangeInput) (SetJobBudgetRangeOutput, error) {
	if in.JobID <= 0 {
		return SetJobBudgetRangeOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return SetJobBudgetRangeOutput{}, fmt.Errorf("client_id is required")
	}
	if in.BudgetMin <= 0 {
		return SetJobBudgetRangeOutput{}, fmt.Errorf("budget_min must be greater than zero")
	}
	if in.BudgetMax < in.BudgetMin {
		return SetJobBudgetRangeOutput{}, fmt.Errorf("budget_max must be greater than or equal to budget_min")
	}

	job, err := uc.Jobs.SetBudgetRange(ctx, in.JobID, in.ClientID, in.BudgetMin, in.BudgetMax, uc.Clock.Now())
	if err != nil {
		return SetJobBudgetRangeOutput{}, err
	}
	return SetJobBudgetRangeOutput{Job: job}, nil
}

type SetJobExperienceLevel struct {
	Jobs  JobRepository
	Clock Clock
}

type SetJobExperienceLevelInput struct {
	JobID           int64
	ClientID        uuid.UUID
	ExperienceLevel string
}

type SetJobExperienceLevelOutput struct {
	Job domain.Job
}

func (uc *SetJobExperienceLevel) Execute(ctx context.Context, in SetJobExperienceLevelInput) (SetJobExperienceLevelOutput, error) {
	if in.JobID <= 0 {
		return SetJobExperienceLevelOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return SetJobExperienceLevelOutput{}, fmt.Errorf("client_id is required")
	}
	level, err := domain.ValidateExperienceLevel(in.ExperienceLevel)
	if err != nil {
		return SetJobExperienceLevelOutput{}, err
	}
	if level == "" {
		return SetJobExperienceLevelOutput{}, fmt.Errorf("experience_level is required")
	}

	job, err := uc.Jobs.SetExperienceLevel(ctx, in.JobID, in.ClientID, level, uc.Clock.Now())
	if err != nil {
		return SetJobExperienceLevelOutput{}, err
	}
	return SetJobExperienceLevelOutput{Job: job}, nil
}
