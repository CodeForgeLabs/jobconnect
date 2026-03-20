package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"jobconnect/verification/internal/domain"

	"github.com/google/uuid"
)

type RequestReverificationInput struct {
	UserID          uuid.UUID
	ReviewerUserID  uuid.UUID
	Reason          string
	ReverifyDueAt   time.Time
}

type RequestReverification struct {
	Repo  VerificationRepository
	Clock Clock
}

func (uc *RequestReverification) Execute(ctx context.Context, in RequestReverificationInput) (domain.VerificationRequest, error) {
	if in.UserID == uuid.Nil {
		return domain.VerificationRequest{}, fmt.Errorf("user_id is required")
	}
	if in.ReviewerUserID == uuid.Nil {
		return domain.VerificationRequest{}, fmt.Errorf("reviewer_user_id is required")
	}
	if strings.TrimSpace(in.Reason) == "" {
		return domain.VerificationRequest{}, fmt.Errorf("reason is required")
	}
	if in.ReverifyDueAt.IsZero() {
		return domain.VerificationRequest{}, fmt.Errorf("reverify_due_at is required")
	}

	now := uc.Clock.Now()
	out, err := uc.Repo.MarkReverificationRequired(ctx, in.UserID, in.ReviewerUserID, in.Reason, in.ReverifyDueAt.UTC(), now)
	if err != nil {
		return domain.VerificationRequest{}, err
	}

	details, _ := json.Marshal(map[string]any{"status": out.Status, "reason": in.Reason, "reverify_due_at_unix": in.ReverifyDueAt.UTC().Unix()})
	reviewer := in.ReviewerUserID
	_ = uc.Repo.AppendEvent(ctx, domain.VerificationEvent{
		RequestID:   out.ID,
		UserID:      out.UserID,
		EventType:   "reverification_requested",
		ActorUserID: &reviewer,
		DetailsJSON: string(details),
		CreatedAt:   now,
	})

	return out, nil
}
