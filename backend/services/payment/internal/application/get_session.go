package application

import (
	"context"
	"fmt"

	"jobconnect/payment/internal/domain"
)

// GetSession retrieves a payment session by ID.
type GetSession struct {
	Sessions PaymentSessionRepository
}

func (uc *GetSession) Execute(ctx context.Context, sessionID int64) (domain.PaymentSession, error) {
	session, err := uc.Sessions.GetByID(ctx, sessionID)
	if err != nil {
		return domain.PaymentSession{}, fmt.Errorf("get session: %w", err)
	}
	return session, nil
}
