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
	StatusRevoked           = "revoked"
	StatusPaused            = "paused"
	StatusEnded             = "ended"

	MilestoneStatusPending                   = "pending"
	MilestoneStatusSubmitted                 = "submitted"
	MilestoneStatusChangesRequested          = "changes_requested"
	MilestoneStatusApproved                  = "approved"
	MilestoneStatusApprovedPendingSettlement = "approved_pending_settlement"
	MilestoneStatusFunded                    = "funded"
	MilestoneStatusReleased                  = "released"

	HourlyLogStatusPending  = "pending"
	HourlyLogStatusApproved = "approved"
	HourlyLogStatusRejected = "rejected"

	HourlyInvoiceStatusDraft     = "draft"
	HourlyInvoiceStatusSubmitted = "submitted"
	HourlyInvoiceStatusInReview  = "in_review"
	HourlyInvoiceStatusApproved  = "approved"
	HourlyInvoiceStatusDisputed  = "disputed"
	HourlyInvoiceStatusCharged   = "charged"
	HourlyInvoiceStatusPaid      = "paid"
	HourlyInvoiceStatusFailed    = "failed"

	ContractBonusStatusPending = "pending"
	ContractBonusStatusPaid    = "paid"
	ContractBonusStatusFailed  = "failed"

	AmendmentStatusPending  = "pending"
	AmendmentStatusAccepted = "accepted"
	AmendmentStatusRejected = "rejected"
	AmendmentStatusExpired  = "expired"

	StatusHistoryEventContractStatusChanged              = "contract_status_changed"
	StatusHistoryEventMilestoneSubmitted                 = "milestone_submitted"
	StatusHistoryEventMilestoneChangesRequested          = "milestone_changes_requested"
	StatusHistoryEventMilestoneApprovedPendingSettlement = "milestone_approved_pending_settlement"
	StatusHistoryEventMilestoneFunded                    = "milestone_funded"
	StatusHistoryEventMilestoneReleased                  = "milestone_released"
	StatusHistoryEventHourlyInvoiceCreated               = "hourly_invoice_created"
	StatusHistoryEventHourlyInvoiceDisputed              = "hourly_invoice_disputed"
	StatusHistoryEventHourlyInvoicePaid                  = "hourly_invoice_paid"
	StatusHistoryEventContractEndBlocked                 = "contract_end_blocked"
)

type Milestone struct {
	ID             int64
	Title          string
	Description    string
	Amount         float64
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
	HourlyRate      float64
	FixedTotal      float64
	WeeklyHourLimit int32

	Milestones []Milestone

	CreatedAt   time.Time
	UpdatedAt   time.Time
	ActivatedAt *time.Time
	DeclinedAt  *time.Time
	RevokedAt   *time.Time
	PausedAt    *time.Time
	EndedAt     *time.Time
}

type JobOfferState struct {
	JobID             int64
	HasPendingOffer   bool
	PendingContractID int64
	HasActiveContract bool
	ActiveContractID  int64
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
	EvidenceURLs   []string
	Status         string
	ReviewNote     string
	CreatedAt      time.Time
	ClientReviewAt *time.Time
	InvoiceID      int64
}

type HourlyWorkSummary struct {
	ContractID              int64
	WeekStart               time.Time
	WeekEnd                 time.Time
	WeeklyHourLimit         int32
	BillableMinutes         int32
	PendingMinutes          int32
	ApprovedMinutes         int32
	RejectedMinutes         int32
	RemainingMinutes        int32
	HourlyRate              float64
	EstimatedBillableAmount float64
}

type HourlyInvoice struct {
	ID              int64
	ContractID      int64
	ClientID        uuid.UUID
	FreelancerID    uuid.UUID
	WeekStart       time.Time
	WeekEnd         time.Time
	Status          string
	BillableMinutes int32
	HourlyRate      float64
	AmountMinor     int64
	DisputeID       string
	CreatedAt       time.Time
	SubmittedAt     *time.Time
	ApprovedAt      *time.Time
	PaidAt          *time.Time
	FailedAt        *time.Time
}

type ContractBonus struct {
	ID                 int64
	ContractID         int64
	ClientID           uuid.UUID
	FreelancerID       uuid.UUID
	AmountMinor        int64
	PaymentReferenceID string
	Note               string
	Status             string
	CreatedAt          time.Time
	PaidAt             *time.Time
	FailedAt           *time.Time
}

type Amendment struct {
	ID           int64
	ContractID   int64
	ProposedBy   uuid.UUID
	Summary      string
	Payload      AmendmentPayload
	Status       string
	ExpiresAt    *time.Time
	RespondedAt  *time.Time
	CreatedAt    time.Time
	RespondedBy  *uuid.UUID
	ResponseNote string
}

type CompensationChange struct {
	NewHourlyRate float64
	NewFixedTotal float64
}

type MilestonesChange struct {
	Milestones []Milestone
}

type WeeklyLimitChange struct {
	NewWeeklyHourLimit int32
}

type ScopeChange struct {
	NewTitle       string
	NewDescription string
}

type AmendmentPayload struct {
	CompensationChange *CompensationChange
	MilestonesChange   *MilestonesChange
	WeeklyLimitChange  *WeeklyLimitChange
	ScopeChange        *ScopeChange
}

type StatusHistoryEntry struct {
	ID          int64
	ContractID  int64
	Status      string
	Reason      string
	ActorID     uuid.UUID
	CreatedAt   time.Time
	EventType   string
	MilestoneID int64
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
	case StatusPendingAcceptance, StatusActive, StatusDeclined, StatusRevoked, StatusPaused, StatusEnded:
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
	if strings.EqualFold(c.ContractType, TypeFixed) {
		if _, err := MoneyToMinorUnits(c.FixedTotal, "fixed_total"); err != nil {
			return err
		}
		if c.FixedTotal <= 0 {
			return fmt.Errorf("fixed_total must be greater than zero")
		}
		if _, err := NormalizeMilestonesForContract(c.ContractType, c.FixedTotal, c.Milestones); err != nil {
			return err
		}
	}
	if strings.EqualFold(c.ContractType, TypeHourly) {
		if _, err := MoneyToMinorUnits(c.HourlyRate, "hourly_rate"); err != nil {
			return err
		}
		if c.HourlyRate <= 0 {
			return fmt.Errorf("hourly_rate must be greater than zero")
		}
		if len(c.Milestones) > 0 {
			return fmt.Errorf("hourly contracts cannot include milestones")
		}
	}
	return nil
}

func NormalizeMilestonesForContract(contractType string, fixedTotal float64, milestones []Milestone) ([]Milestone, error) {
	contractType = strings.ToLower(strings.TrimSpace(contractType))
	switch contractType {
	case TypeHourly:
		if len(milestones) > 0 {
			return nil, fmt.Errorf("hourly contracts cannot include milestones")
		}
		return nil, nil
	case TypeFixed:
	default:
		return nil, fmt.Errorf("invalid contract_type")
	}

	if len(milestones) == 0 {
		return nil, fmt.Errorf("fixed contracts require at least one milestone")
	}
	if fixedTotal <= 0 {
		return nil, fmt.Errorf("fixed_total must be greater than zero")
	}

	out := make([]Milestone, 0, len(milestones))
	var totalMinor int64
	for i, m := range milestones {
		title := strings.TrimSpace(m.Title)
		if title == "" {
			return nil, fmt.Errorf("milestone title is required at index %d", i)
		}
		if len(title) > 200 {
			return nil, fmt.Errorf("milestone title too long at index %d", i)
		}
		if m.Amount <= 0 {
			return nil, fmt.Errorf("milestone amount must be positive at index %d", i)
		}
		if m.DueAt == nil || m.DueAt.IsZero() {
			return nil, fmt.Errorf("milestone due_at is required at index %d", i)
		}
		amountMinor, err := MoneyToMinorUnits(m.Amount, fmt.Sprintf("milestone amount at index %d", i))
		if err != nil {
			return nil, err
		}
		if amountMinor <= 0 {
			return nil, fmt.Errorf("milestone amount must be positive at index %d", i)
		}
		totalMinor += amountMinor
		dueAt := m.DueAt.UTC()
		out = append(out, Milestone{
			ID:          int64(i + 1),
			Title:       title,
			Description: strings.TrimSpace(m.Description),
			Amount:      m.Amount,
			DueAt:       &dueAt,
			Status:      MilestoneStatusPending,
		})
	}

	expectedMinor, err := MoneyToMinorUnits(fixedTotal, "fixed_total")
	if err != nil {
		return nil, err
	}
	if totalMinor != expectedMinor {
		return nil, fmt.Errorf("milestone amounts must equal fixed_total")
	}
	return out, nil
}

func CanAccept(current string) bool {
	return strings.EqualFold(strings.TrimSpace(current), StatusPendingAcceptance)
}

func CanDecline(current string) bool {
	return strings.EqualFold(strings.TrimSpace(current), StatusPendingAcceptance)
}

func CanRevoke(current string) bool {
	return strings.EqualFold(strings.TrimSpace(current), StatusPendingAcceptance)
}

func ValidateMilestoneStatus(v string) error {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case MilestoneStatusPending, MilestoneStatusSubmitted, MilestoneStatusChangesRequested, MilestoneStatusApproved, MilestoneStatusApprovedPendingSettlement, MilestoneStatusFunded, MilestoneStatusReleased:
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
