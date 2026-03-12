package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	TypeFixed  = "fixed"
	TypeHourly = "hourly"

	StatusPendingAcceptance = "pending_acceptance"
	StatusActive            = "active"
	StatusDeclined          = "declined"
	StatusPaused            = "paused"
	StatusEnded             = "ended"

	MilestoneStatusPending          = "pending"
	MilestoneStatusSubmitted        = "submitted"
	MilestoneStatusChangesRequested = "changes_requested"
	MilestoneStatusApproved         = "approved"

	HourlyLogStatusPending  = "pending"
	HourlyLogStatusApproved = "approved"
	HourlyLogStatusRejected = "rejected"

	AmendmentStatusPending  = "pending"
	AmendmentStatusAccepted = "accepted"
	AmendmentStatusRejected = "rejected"
	AmendmentStatusExpired  = "expired"
)

type Milestone struct {
	ID             int64
	Title          string
	Description    string
	Amount         float64
	Currency       string
	DueAt          *time.Time
	Status         string
	SubmissionNote string
	SubmissionURLs []string
	SubmittedAt    *time.Time
	ReviewNote     string
	ReviewedAt     *time.Time
	RevisionCount  int32
}

type Contract struct {
	ID           int64
	ClientID     uuid.UUID
	FreelancerID uuid.UUID
	JobID        int64
	ProposalID   int64

	ContractType string
	Status       string

	Title           string
	Description     string
	Currency        string
	HourlyRate      float64
	FixedTotal      float64
	WeeklyHourLimit int32

	Milestones []Milestone

	CreatedAt   time.Time
	UpdatedAt   time.Time
	ActivatedAt *time.Time
	DeclinedAt  *time.Time
	PausedAt    *time.Time
	EndedAt     *time.Time
}

type HourlyLog struct {
	ID             int64
	ContractID     int64
	FreelancerID   uuid.UUID
	WorkDate       time.Time
	StartAt        time.Time
	EndAt          time.Time
	DurationMin    int32
	Note           string
	Status         string
	ReviewNote     string
	CreatedAt      time.Time
	ClientReviewAt *time.Time
}

type Amendment struct {
	ID          int64
	ContractID  int64
	ProposedBy  uuid.UUID
	Summary     string
	PayloadJSON string
	Status      string
	ExpiresAt   *time.Time
	RespondedAt *time.Time
	CreatedAt   time.Time
}

type StatusHistoryEntry struct {
	ID         int64
	ContractID int64
	Status     string
	Reason     string
	ActorID    uuid.UUID
	CreatedAt  time.Time
}

func ValidateType(v string) error {
	t := strings.ToLower(strings.TrimSpace(v))
	switch t {
	case TypeFixed, TypeHourly:
		return nil
	default:
		return fmt.Errorf("invalid contract_type")
	}
}

func ValidateStatus(v string) error {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case StatusPendingAcceptance, StatusActive, StatusDeclined, StatusPaused, StatusEnded:
		return nil
	default:
		return fmt.Errorf("invalid contract_status")
	}
}

func ValidateForCreate(c Contract) error {
	if c.ClientID == uuid.Nil {
		return fmt.Errorf("client_id is required")
	}
	if c.FreelancerID == uuid.Nil {
		return fmt.Errorf("freelancer_id is required")
	}
	if strings.TrimSpace(c.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if len(strings.TrimSpace(c.Title)) > 200 {
		return fmt.Errorf("title too long")
	}
	if err := ValidateType(c.ContractType); err != nil {
		return err
	}
	if strings.TrimSpace(c.Currency) == "" {
		return fmt.Errorf("currency is required")
	}

	if strings.EqualFold(c.ContractType, TypeFixed) {
		if c.FixedTotal <= 0 {
			return fmt.Errorf("fixed_total must be greater than zero")
		}
	}
	if strings.EqualFold(c.ContractType, TypeHourly) {
		if c.HourlyRate <= 0 {
			return fmt.Errorf("hourly_rate must be greater than zero")
		}
	}
	return nil
}

func CanAccept(current string) bool {
	return strings.EqualFold(strings.TrimSpace(current), StatusPendingAcceptance)
}

func CanDecline(current string) bool {
	return strings.EqualFold(strings.TrimSpace(current), StatusPendingAcceptance)
}

func ValidateMilestoneStatus(v string) error {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case MilestoneStatusPending, MilestoneStatusSubmitted, MilestoneStatusChangesRequested, MilestoneStatusApproved:
		return nil
	default:
		return fmt.Errorf("invalid milestone_status")
	}
}

func ValidateHourlyLogStatus(v string) error {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case HourlyLogStatusPending, HourlyLogStatusApproved, HourlyLogStatusRejected:
		return nil
	default:
		return fmt.Errorf("invalid hourly_log_status")
	}
}

func ValidateAmendmentStatus(v string) error {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case AmendmentStatusPending, AmendmentStatusAccepted, AmendmentStatusRejected, AmendmentStatusExpired:
		return nil
	default:
		return fmt.Errorf("invalid amendment_status")
	}
}
