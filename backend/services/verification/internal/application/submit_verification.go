package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"jobconnect/verification/internal/domain"

	"github.com/google/uuid"
)

type SubmitVerificationInput struct {
	UserID               uuid.UUID
	LegalName            string
	CountryCode          string
	DocumentType         string
	DocumentNumberMasked string
	EvidenceURLs         []string
	SubmissionNote       string
}

type SubmitVerification struct {
	Repo  VerificationRepository
	Clock Clock
}

func (uc *SubmitVerification) Execute(ctx context.Context, in SubmitVerificationInput) (domain.VerificationRequest, error) {
	if in.UserID == uuid.Nil {
		return domain.VerificationRequest{}, fmt.Errorf("user_id is required")
	}
	if err := domain.ValidateSubmission(in.LegalName, in.CountryCode, in.DocumentType, in.DocumentNumberMasked); err != nil {
		return domain.VerificationRequest{}, err
	}

	version := int32(1)
	latest, err := uc.Repo.GetLatestByUserID(ctx, in.UserID)
	if err == nil {
		if latest.Status == domain.StatusPendingReview {
			return domain.VerificationRequest{}, fmt.Errorf("verification request already pending")
		}
		version = latest.RequestVersion + 1
	}

	now := uc.Clock.Now()
	created, err := uc.Repo.CreateSubmission(ctx, domain.VerificationRequest{
		UserID:               in.UserID,
		RequestVersion:       version,
		Status:               domain.StatusPendingReview,
		LegalName:            strings.TrimSpace(in.LegalName),
		CountryCode:          strings.ToUpper(strings.TrimSpace(in.CountryCode)),
		DocumentType:         strings.TrimSpace(in.DocumentType),
		DocumentNumberMasked: strings.TrimSpace(in.DocumentNumberMasked),
		EvidenceURLs:         in.EvidenceURLs,
		SubmissionNote:       strings.TrimSpace(in.SubmissionNote),
		SubmittedAt:          now,
		UpdatedAt:            now,
	})
	if err != nil {
		return domain.VerificationRequest{}, err
	}

	details, _ := json.Marshal(map[string]any{"status": created.Status, "version": created.RequestVersion})
	_ = uc.Repo.AppendEvent(ctx, domain.VerificationEvent{
		RequestID:   created.ID,
		UserID:      created.UserID,
		EventType:   "submitted",
		DetailsJSON: string(details),
		CreatedAt:   now,
	})

	return created, nil
}
