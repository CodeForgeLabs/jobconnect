package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	StatusOpen     = "open"
	StatusResolved = "resolved"

	DecisionRelease = "release"
	DecisionRefund  = "refund"
)

type Dispute struct {
	ID             int64
	ReferenceType  string
	ReferenceID    string
	OpenedBy       uuid.UUID
	Reason         string
	Status         string
	Decision       string
	ResolutionNote string
	ResolvedBy     *uuid.UUID
	CreatedAt      time.Time
	ResolvedAt     *time.Time
}

func ValidateOpen(referenceType, referenceID, reason string) error {
	if strings.TrimSpace(referenceType) == "" || strings.TrimSpace(referenceID) == "" {
		return fmt.Errorf("reference_type and reference_id are required")
	}
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("reason is required")
	}
	return nil
}

func ValidateDecision(v string) error {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case DecisionRelease, DecisionRefund:
		return nil
	default:
		return fmt.Errorf("invalid decision")
	}
}
