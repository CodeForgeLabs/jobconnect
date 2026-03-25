package application

import (
	"context"
	"fmt"

	"jobconnect/payment/internal/domain"

	"github.com/google/uuid"
)

// InitiateDeposit creates a PENDING payment session and returns a checkout URL.
type InitiateDeposit struct {
	Sessions PaymentSessionRepository
	Gateway  func(provider string) PaymentGateway // factory to pick provider
	Clock    Clock
}

// InitiateDepositInput is the input for initiating a deposit.
type InitiateDepositInput struct {
	UserID        uuid.UUID
	Provider      string
	AmountMinor   int64
	Currency      string
	ReferenceType string // "milestone"
	ReferenceID   string // milestone ID
	ReturnURL     string
}

func (uc *InitiateDeposit) Execute(ctx context.Context, in InitiateDepositInput) (domain.PaymentSession, string, error) {
	if err := domain.ValidateDepositInput(in.Provider, in.AmountMinor, in.ReferenceType, in.ReferenceID); err != nil {
		return domain.PaymentSession{}, "", err
	}

	now := uc.Clock.Now()
	idempotencyKey := fmt.Sprintf("dep_%s_%s_%s", in.UserID, in.ReferenceType, in.ReferenceID)

	// Idempotency check: return existing session if found.
	if existing, found, err := uc.Sessions.GetByIdempotencyKey(ctx, idempotencyKey); err != nil {
		return domain.PaymentSession{}, "", fmt.Errorf("idempotency check: %w", err)
	} else if found && existing.Status == domain.StatusPending {
		return existing, existing.ExternalRef, nil
	}

	currency := in.Currency
	if currency == "" {
		currency = domain.DefaultCurrency
	}

	session := domain.PaymentSession{
		UserID:         in.UserID,
		Provider:       in.Provider,
		PaymentType:    domain.TypeDeposit,
		Status:         domain.StatusPending,
		AmountMinor:    in.AmountMinor,
		Currency:       currency,
		IdempotencyKey: idempotencyKey,
		ReferenceType:  in.ReferenceType,
		ReferenceID:    in.ReferenceID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	gw := uc.Gateway(in.Provider)
	result, err := gw.InitiateCheckout(ctx, CheckoutInput{
		AmountMinor: in.AmountMinor,
		Currency:    currency,
		TxRef:       idempotencyKey,
		ReturnURL:   in.ReturnURL,
	})
	if err != nil {
		return domain.PaymentSession{}, "", fmt.Errorf("gateway checkout: %w", err)
	}

	session.ExternalRef = result.ExternalRef

	created, err := uc.Sessions.Create(ctx, session)
	if err != nil {
		return domain.PaymentSession{}, "", fmt.Errorf("create session: %w", err)
	}

	return created, result.CheckoutURL, nil
}
