package application

import (
	"context"
	"fmt"

	"jobconnect/payment/internal/domain"

	"github.com/google/uuid"
)

// UploadReceipt handles uploading a payment receipt to MinIO.
type UploadReceipt struct {
	Sessions PaymentSessionRepository
	Receipts ReceiptStore
	Clock    Clock
}

// UploadReceiptInput is the input for uploading a receipt.
type UploadReceiptInput struct {
	UserID      uuid.UUID
	SessionID   int64
	ReceiptData []byte
	ContentType string
}

func (uc *UploadReceipt) Execute(ctx context.Context, in UploadReceiptInput) (domain.PaymentSession, string, error) {
	session, err := uc.Sessions.GetByID(ctx, in.SessionID)
	if err != nil {
		return domain.PaymentSession{}, "", fmt.Errorf("get session: %w", err)
	}

	// Only the session owner can upload a receipt.
	if session.UserID != in.UserID {
		return domain.PaymentSession{}, "", fmt.Errorf("%w: session does not belong to user", domain.ErrUnauthorized)
	}

	// Only upload for pending sessions.
	if session.Status != domain.StatusPending {
		return domain.PaymentSession{}, "", fmt.Errorf("%w: can only upload receipt for pending sessions", domain.ErrInvalidArgument)
	}

	if len(in.ReceiptData) == 0 {
		return domain.PaymentSession{}, "", fmt.Errorf("%w: receipt data is empty", domain.ErrInvalidArgument)
	}

	// Generate storage key.
	key := fmt.Sprintf("receipts/%d/%s", in.SessionID, uuid.New().String())

	if err := uc.Receipts.PutReceipt(ctx, key, in.ReceiptData, in.ContentType); err != nil {
		return domain.PaymentSession{}, "", fmt.Errorf("store receipt: %w", err)
	}

	session.ReceiptKey = key
	session.UpdatedAt = uc.Clock.Now()
	if err := uc.Sessions.Update(ctx, session); err != nil {
		return domain.PaymentSession{}, "", fmt.Errorf("update session: %w", err)
	}

	return session, key, nil
}
