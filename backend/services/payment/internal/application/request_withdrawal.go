package application

import (
	"context"
	"fmt"

	"jobconnect/payment/internal/domain"

	"github.com/google/uuid"
)

// RequestWithdrawal handles a freelancer's withdrawal request.
type RequestWithdrawal struct {
	Sessions     PaymentSessionRepository
	Gateway      func(provider string) PaymentGateway
	Wallet       WalletClient
	Verification VerificationClient
	Clock        Clock
}

// RequestWithdrawalInput is the input for requesting a withdrawal.
type RequestWithdrawalInput struct {
	UserID            uuid.UUID
	Provider          string
	AmountMinor       int64
	BankCode          string
	AccountNumber     string
	AccountHolderName string
}

func (uc *RequestWithdrawal) Execute(ctx context.Context, in RequestWithdrawalInput) (domain.PaymentSession, error) {
	if err := domain.ValidateWithdrawalInput(in.Provider, in.AmountMinor, in.BankCode, in.AccountNumber); err != nil {
		return domain.PaymentSession{}, err
	}

	// 1. KYC check.
	verified, err := uc.Verification.IsKYCVerified(ctx, in.UserID)
	if err != nil {
		return domain.PaymentSession{}, fmt.Errorf("kyc check: %w", err)
	}
	if !verified {
		return domain.PaymentSession{}, fmt.Errorf("%w: complete identity verification before withdrawing", domain.ErrKYCNotVerified)
	}

	now := uc.Clock.Now()
	idempotencyKey := fmt.Sprintf("wd_%s_%d", in.UserID, now.UnixMilli())

	session := domain.PaymentSession{
		UserID:         in.UserID,
		Provider:       in.Provider,
		PaymentType:    domain.TypeWithdrawal,
		Status:         domain.StatusPending,
		AmountMinor:    in.AmountMinor,
		IdempotencyKey: idempotencyKey,
		ReferenceType:  "withdrawal",
		ReferenceID:    in.AccountNumber,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	created, err := uc.Sessions.Create(ctx, session)
	if err != nil {
		return domain.PaymentSession{}, fmt.Errorf("create session: %w", err)
	}

	// 2. Debit the user's wallet.
	debitKey := fmt.Sprintf("debit_%d", created.ID)
	if err := uc.Wallet.DebitWalletInternal(ctx, DebitInput{
		AmountMinor:    in.AmountMinor,
		IdempotencyKey: debitKey,
		ReferenceType:  "payment_session",
		ReferenceID:    fmt.Sprintf("%d", created.ID),
		Note:           "Withdrawal request",
	}); err != nil {
		// Mark session as failed if debit fails.
		_ = created.MarkFailed(now, fmt.Sprintf("wallet debit failed: %v", err))
		_ = uc.Sessions.Update(ctx, created)
		return created, fmt.Errorf("wallet debit: %w", err)
	}

	// 3. Initiate bank transfer.
	gw := uc.Gateway(in.Provider)
	result, err := gw.InitiateTransfer(ctx, TransferInput{
		AmountMinor:       in.AmountMinor,
		BankCode:          in.BankCode,
		AccountNumber:     in.AccountNumber,
		AccountHolderName: in.AccountHolderName,
		TxRef:             idempotencyKey,
	})
	if err != nil {
		// TODO: should we reverse the debit here? For now, mark failed.
		_ = created.MarkFailed(now, fmt.Sprintf("transfer failed: %v", err))
		_ = uc.Sessions.Update(ctx, created)
		return created, fmt.Errorf("gateway transfer: %w", err)
	}

	created.ExternalRef = result.ExternalRef
	_ = uc.Sessions.Update(ctx, created)

	return created, nil
}
