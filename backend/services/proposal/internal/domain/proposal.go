package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	StatusSent        = "sent"
	StatusShortlisted = "shortlisted"
	StatusRejected    = "rejected"
	StatusHired       = "hired"
	StatusWithdrawn   = "withdrawn"

	BidTypeFixed  = "fixed"
	BidTypeHourly = "hourly"

	SortNewest  = "newest"
	SortOldest  = "oldest"
	SortBidHigh = "bid_high"
	SortBidLow  = "bid_low"
)

type Attachment struct {
	ID          int64
	FileName    string
	ContentType string
	URL         string
	SizeBytes   int64
	StorageKey  string
}

type Proposal struct {
	ID           int64
	JobID        int64
	ClientID     uuid.UUID
	FreelancerID uuid.UUID

	CoverLetter   string
	BidType       string
	BidAmount     float64
	EstimatedDays int32
	Attachments   []Attachment

	Status       string
	StatusReason string

	CreatedAt     time.Time
	UpdatedAt     time.Time
	ShortlistedAt *time.Time
	RejectedAt    *time.Time
	HiredAt       *time.Time
	WithdrawnAt   *time.Time

	ConnectsSpent int32
}

func ValidateBidType(v string) error {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case BidTypeFixed, BidTypeHourly:
		return nil
	default:
		return fmt.Errorf("invalid bid_type")
	}
}

func ValidateStatus(v string) error {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case StatusSent, StatusShortlisted, StatusRejected, StatusHired, StatusWithdrawn:
		return nil
	default:
		return fmt.Errorf("invalid status")
	}
}

func ValidateSortBy(v string) error {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == "" {
		return nil
	}
	switch s {
	case SortNewest, SortOldest, SortBidHigh, SortBidLow:
		return nil
	default:
		return fmt.Errorf("invalid sort_by")
	}
}

func ValidateForSubmit(p Proposal) error {
	if p.JobID <= 0 {
		return fmt.Errorf("job_id is required")
	}
	if p.ClientID == uuid.Nil {
		return fmt.Errorf("client_id is required")
	}
	if p.FreelancerID == uuid.Nil {
		return fmt.Errorf("freelancer_id is required")
	}
	if strings.TrimSpace(p.CoverLetter) == "" {
		return fmt.Errorf("cover_letter is required")
	}
	if len(strings.TrimSpace(p.CoverLetter)) > 5000 {
		return fmt.Errorf("cover_letter too long")
	}
	if err := ValidateBidType(p.BidType); err != nil {
		return err
	}
	if p.BidAmount <= 0 {
		return fmt.Errorf("bid_amount must be greater than zero")
	}
	if p.EstimatedDays <= 0 {
		return fmt.Errorf("estimated_days must be greater than zero")
	}
	if len(p.Attachments) > 20 {
		return fmt.Errorf("too many attachments")
	}
	for _, a := range p.Attachments {
		if strings.TrimSpace(a.FileName) == "" {
			return fmt.Errorf("attachment file_name is required")
		}
		if strings.TrimSpace(a.ContentType) == "" {
			return fmt.Errorf("attachment content_type is required")
		}
		if strings.TrimSpace(a.URL) == "" && strings.TrimSpace(a.StorageKey) == "" {
			return fmt.Errorf("attachment url or storage_key is required")
		}
		if a.SizeBytes < 0 {
			return fmt.Errorf("attachment size_bytes must be non-negative")
		}
	}
	return nil
}

func ValidateForModify(coverLetter string, bidAmount float64, estimatedDays int32, attachments []Attachment) error {
	if strings.TrimSpace(coverLetter) == "" {
		return fmt.Errorf("cover_letter is required")
	}
	if len(strings.TrimSpace(coverLetter)) > 5000 {
		return fmt.Errorf("cover_letter too long")
	}
	if bidAmount <= 0 {
		return fmt.Errorf("bid_amount must be greater than zero")
	}
	if estimatedDays <= 0 {
		return fmt.Errorf("estimated_days must be greater than zero")
	}
	if len(attachments) > 20 {
		return fmt.Errorf("too many attachments")
	}
	return nil
}

func CanFreelancerModify(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), StatusSent)
}

func CanFreelancerWithdraw(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))
	return s == StatusSent || s == StatusShortlisted
}

func CanTransition(currentStatus, nextStatus string) bool {
	current := strings.ToLower(strings.TrimSpace(currentStatus))
	next := strings.ToLower(strings.TrimSpace(nextStatus))

	switch current {
	case StatusSent:
		return next == StatusShortlisted || next == StatusRejected || next == StatusHired || next == StatusWithdrawn
	case StatusShortlisted:
		return next == StatusRejected || next == StatusHired || next == StatusWithdrawn
	default:
		return false
	}
}
