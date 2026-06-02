package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"jobconnect/verification/internal/domain"

	"github.com/google/uuid"
)

type ReviewVerificationInput struct {
	RequestID       int64
	ReviewerUserID  uuid.UUID
	Decision        string
	RejectionReason string
	InternalNote    string
}

type ReviewVerification struct {
	Repo  VerificationRepository
	Clock Clock
}

func (uc *ReviewVerification) Execute(ctx context.Context, in ReviewVerificationInput) (domain.VerificationRequest, error) {
	if in.RequestID <= 0 {
		return domain.VerificationRequest{}, fmt.Errorf("request_id is required")
	}
	if in.ReviewerUserID == uuid.Nil {
		return domain.VerificationRequest{}, fmt.Errorf("reviewer_user_id is required")
	}
	if err := domain.ValidateDecision(in.Decision, in.RejectionReason); err != nil {
		return domain.VerificationRequest{}, err
	}

	now := uc.Clock.Now()
	out, err := uc.Repo.Review(ctx, in.RequestID, in.ReviewerUserID, strings.ToLower(strings.TrimSpace(in.Decision)), in.RejectionReason, in.InternalNote, now)
	if err != nil {
		return domain.VerificationRequest{}, err
	}

	details, _ := json.Marshal(map[string]any{"status": out.Status, "decision": in.Decision})
	reviewer := in.ReviewerUserID
	_ = uc.Repo.AppendEvent(ctx, domain.VerificationEvent{
		RequestID:   out.ID,
		UserID:      out.UserID,
		EventType:   "reviewed",
		ActorUserID: &reviewer,
		DetailsJSON: string(details),
		CreatedAt:   now,
	})

	return out, nil
}
