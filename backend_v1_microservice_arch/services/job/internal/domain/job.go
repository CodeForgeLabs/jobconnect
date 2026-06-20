package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	RoleClient = "client"

	JobTypeFixed  = "fixed"
	JobTypeHourly = "hourly"

	JobStatusOpen      = "open"
	JobStatusPaused    = "paused"
	JobStatusFilled    = "filled"
	JobStatusClosed    = "closed"
	JobStatusCompleted = "completed"
	JobStatusCanceled  = "canceled"

	VisibilityPublic     = "public"
	VisibilityPrivate    = "private"
	VisibilityInviteOnly = "invite_only"

	CloseReasonCanceled = "canceled"
)

type Attachment struct {
	ID          int64
	FileName    string
	ContentType string
	StorageKey  string
	URL         string
	SizeBytes   int64
}

type Job struct {
	ID                 int64
	ClientID           uuid.UUID
	Title              string
	Description        string
	RequiredSkills     []string
	JobType            string
	BudgetFixed        float64
	HourlyRate         float64
	BudgetMin          float64
	BudgetMax          float64
	Visibility         string
	Deadline           *time.Time
	Attachments        []Attachment
	Status             string
	CloseReason        string
	SettlementPolicy   string
	CancellationReason string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	ClosedAt           *time.Time
	PausedAt           *time.Time
	FilledAt           *time.Time
	CompletedAt        *time.Time
	CanceledAt         *time.Time
}

type JobInvite struct {
	JobID          int64
	ClientID       uuid.UUID
	FreelancerID   uuid.UUID
	InvitedAt      time.Time
	ResponseStatus string
}

type InvitedJob struct {
	Job    Job
	Invite JobInvite
}

func ValidateJobType(jobType string) (string, error) {
	jobType = strings.ToLower(strings.TrimSpace(jobType))
	switch jobType {
	case JobTypeFixed, JobTypeHourly:
		return jobType, nil
	default:
		return "", fmt.Errorf("invalid job_type")
	}
}

func ValidateStatus(status string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return "", nil
	}
	switch status {
	case JobStatusOpen, JobStatusClosed, JobStatusPaused, JobStatusFilled, JobStatusCompleted, JobStatusCanceled:
		return status, nil
	default:
		return "", fmt.Errorf("invalid status")
	}
}

func ValidateVisibility(visibility string) (string, error) {
	visibility = strings.ToLower(strings.TrimSpace(visibility))
	if visibility == "" {
		return "", nil
	}
	switch visibility {
	case VisibilityPublic, VisibilityPrivate, VisibilityInviteOnly:
		return visibility, nil
	default:
		return "", fmt.Errorf("invalid visibility")
	}
}

func ValidateCreate(job Job, now time.Time) error {
	if job.ClientID == uuid.Nil {
		return fmt.Errorf("client_id is required")
	}
	job.Title = strings.TrimSpace(job.Title)
	if job.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(job.Title) > 160 {
		return fmt.Errorf("title too long")
	}
	job.Description = strings.TrimSpace(job.Description)
	if job.Description == "" {
		return fmt.Errorf("description is required")
	}
	if len(job.Description) > 10000 {
		return fmt.Errorf("description too long")
	}
	if _, err := ValidateJobType(job.JobType); err != nil {
		return err
	}
	if job.JobType == JobTypeFixed {
		if job.BudgetFixed <= 0 {
			return fmt.Errorf("budget_fixed must be greater than zero")
		}
		if job.HourlyRate > 0 {
			return fmt.Errorf("hourly_rate must be empty for fixed jobs")
		}
	}
	if job.JobType == JobTypeHourly {
		if job.HourlyRate <= 0 {
			return fmt.Errorf("hourly_rate must be greater than zero")
		}
		if job.BudgetFixed > 0 {
			return fmt.Errorf("budget_fixed must be empty for hourly jobs")
		}
	}
	if job.Deadline != nil && !job.Deadline.After(now) {
		return fmt.Errorf("deadline must be in the future")
	}
	if len(job.RequiredSkills) > 100 {
		return fmt.Errorf("too many required_skills")
	}
	for _, s := range job.RequiredSkills {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("required_skills contains empty value")
		}
	}
	if len(job.Attachments) > 20 {
		return fmt.Errorf("too many attachments")
	}
	for _, a := range job.Attachments {
		if strings.TrimSpace(a.FileName) == "" {
			return fmt.Errorf("attachment file_name is required")
		}
		if strings.TrimSpace(a.ContentType) == "" {
			return fmt.Errorf("attachment content_type is required")
		}
		if strings.TrimSpace(a.URL) == "" {
			return fmt.Errorf("attachment url is required")
		}
		if a.SizeBytes < 0 {
			return fmt.Errorf("attachment size_bytes must be non-negative")
		}
	}
	return nil
}

func ValidateCloseReason(reason string) error {
	if len(strings.TrimSpace(reason)) > 500 {
		return fmt.Errorf("close reason too long")
	}
	return nil
}
