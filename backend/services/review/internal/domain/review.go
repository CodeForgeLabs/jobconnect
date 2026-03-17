package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	RoleClient     = "client"
	RoleFreelancer = "freelancer"

	MinRating             = 1
	MaxRating             = 5
	ReviewEditGracePeriod = 24 * time.Hour
)

type Review struct {
	ID           int64
	ContractID   int64
	ReviewerID   uuid.UUID
	RevieweeID   uuid.UUID
	ReviewerRole string
	Rating       int32
	Title        string
	Comment      string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	ReplyComment *string
	RepliedAt    *time.Time
}

func ValidateReviewerRole(role string) error {
	role = strings.ToLower(strings.TrimSpace(role))
	switch role {
	case RoleClient, RoleFreelancer:
		return nil
	default:
		return fmt.Errorf("invalid reviewer_role %q", role)
	}
}

func (r Review) IsWithinGracePeriod(now time.Time) bool {
	return now.Sub(r.CreatedAt) <= ReviewEditGracePeriod
}

func ValidateReply(reply string) error {
	reply = strings.TrimSpace(reply)
	if reply == "" {
		return fmt.Errorf("reply cannot be empty")
	}
	if len(reply) > 5000 {
		return fmt.Errorf("reply too long")
	}
	return nil
}

func ValidateCreate(r Review) error {
	if r.ContractID <= 0 {
		return fmt.Errorf("contract_id is required")
	}
	if r.ReviewerID == uuid.Nil {
		return fmt.Errorf("reviewer_id is required")
	}
	if r.RevieweeID == uuid.Nil {
		return fmt.Errorf("reviewee_id is required")
	}
	if r.ReviewerID == r.RevieweeID {
		return fmt.Errorf("reviewer and reviewee must be different users")
	}
	if r.Rating < MinRating || r.Rating > MaxRating {
		return fmt.Errorf("rating must be between %d and %d", MinRating, MaxRating)
	}
	r.Title = strings.TrimSpace(r.Title)
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(r.Title) > 160 {
		return fmt.Errorf("title too long")
	}
	r.Comment = strings.TrimSpace(r.Comment)
	if r.Comment == "" {
		return fmt.Errorf("comment is required")
	}
	if len(r.Comment) > 5000 {
		return fmt.Errorf("comment too long")
	}
	if err := ValidateReviewerRole(r.ReviewerRole); err != nil {
		return err
	}
	return nil
}
