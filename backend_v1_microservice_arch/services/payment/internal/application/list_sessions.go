package application

import (
	"context"
	"fmt"

	"jobconnect/payment/internal/domain"

	"github.com/google/uuid"
)

// ListSessions lists payment sessions for a user with optional filters.
type ListSessions struct {
	Sessions PaymentSessionRepository
}

// ListSessionsInput is the input for listing sessions.
type ListSessionsInput struct {
	UserID      uuid.UUID
	PaymentType *string
	Status      *string
	PageSize    int
	Offset      int
}

func (uc *ListSessions) Execute(ctx context.Context, in ListSessionsInput) ([]domain.PaymentSession, error) {
	limit := in.PageSize
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	sessions, err := uc.Sessions.ListByUserID(ctx, in.UserID, SessionFilters{
		PaymentType: in.PaymentType,
		Status:      in.Status,
	}, limit, in.Offset)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	return sessions, nil
}
