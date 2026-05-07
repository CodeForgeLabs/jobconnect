package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	StatusUnverified             = "unverified"
	StatusSubmitted              = "submitted"
	StatusPendingReview          = "pending_review"
	StatusVerified               = "verified"
	StatusRejected               = "rejected"
	StatusReverificationRequired = "reverification_required"
)

const (
	DecisionApprove = "approve"
	DecisionReject  = "reject"
)

type VerificationRequest struct {
	ID                   int64
	UserID               uuid.UUID
	RequestVersion       int32
	Status               string
	LegalName            string
	CountryCode          string
	DocumentType         string
	DocumentNumberMasked string
	EvidenceURL          string
	SubmissionNote       string
	ReviewerUserID       *uuid.UUID
	RejectionReason      string
	InternalNote         string
	SubmittedAt          time.Time
	ReviewedAt           *time.Time
	ReverifyDueAt        *time.Time
	UpdatedAt            time.Time
}

type VerificationEvent struct {
	RequestID   int64
	UserID      uuid.UUID
	EventType   string
	ActorUserID *uuid.UUID
	DetailsJSON string
	CreatedAt   time.Time
}

func ValidateSubmission(legalName, countryCode, documentType, masked string) error {
	if strings.TrimSpace(legalName) == "" {
		return fmt.Errorf("legal_name is required")
	}
	if len(strings.TrimSpace(countryCode)) != 2 {
		return fmt.Errorf("country_code must be a 2-letter code")
	}
	if strings.TrimSpace(documentType) == "" {
		return fmt.Errorf("document_type is required")
	}
	if strings.TrimSpace(masked) == "" {
		return fmt.Errorf("document_number_masked is required")
	}
	return nil
}

func ValidateEvidenceURL(evidenceURL string) error {
	if strings.TrimSpace(evidenceURL) == "" {
		return fmt.Errorf("evidence_url is required")
	}
	return nil
}

func ValidateDecision(decision, rejectionReason string) error {
	d := strings.ToLower(strings.TrimSpace(decision))
	switch d {
	case DecisionApprove:
		return nil
	case DecisionReject:
		if strings.TrimSpace(rejectionReason) == "" {
			return fmt.Errorf("rejection_reason is required for reject decision")
		}
		return nil
	default:
		return fmt.Errorf("invalid decision")
	}
}
