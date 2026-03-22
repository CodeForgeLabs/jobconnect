package application

import (
	"context"
	"fmt"

	"jobconnect/payment/internal/domain"
)

// VerifyDeposit is called by the webhook handler or manual poll to confirm a deposit.
// On success it orchestrates: wallet credit → wallet hold → contract milestone funded.
type VerifyDeposit struct {
	Sessions PaymentSessionRepository
	Gateway  func(provider string) PaymentGateway
	Wallet   WalletClient
	Contract ContractClient
	Clock    Clock
}

func (uc *VerifyDeposit) Execute(ctx context.Context, sessionID int64) (domain.PaymentSession, error) {
	session, err := uc.Sessions.GetByID(ctx, sessionID)
	if err != nil {
		return domain.PaymentSession{}, fmt.Errorf("get session: %w", err)
	}

	if session.Status != domain.StatusPending {
		// Already processed — idempotent return.
		return session, nil
	}

	now := uc.Clock.Now()

	// Check expiry.
	if session.IsExpired(now) {
		_ = session.MarkFailed(now, "session expired")
		_ = uc.Sessions.Update(ctx, session)
		return session, fmt.Errorf("%w", domain.ErrSessionExpired)
	}

	// Verify with the payment provider.
	gw := uc.Gateway(session.Provider)
	result, err := gw.VerifyPayment(ctx, session.ExternalRef)
	if err != nil {
		return session, fmt.Errorf("gateway verify: %w", err)
	}

	if !result.Verified {
		_ = session.MarkFailed(now, "payment not verified by provider")
		_ = uc.Sessions.Update(ctx, session)
		return session, nil
	}

	// ── Orchestrate internal state ──

	// 1. Credit the user's wallet.
	creditKey := fmt.Sprintf("credit_%d", session.ID)
	if err := uc.Wallet.CreditWalletInternal(ctx, CreditInput{
		WalletID:       0, // resolved by wallet service via owner_id
		AmountMinor:    session.AmountMinor,
		IdempotencyKey: creditKey,
		ReferenceType:  "payment_session",
		ReferenceID:    fmt.Sprintf("%d", session.ID),
		Note:           fmt.Sprintf("Deposit via %s", session.Provider),
	}); err != nil {
		return session, fmt.Errorf("wallet credit: %w", err)
	}

	// 2. Place hold (escrow) if this is a milestone deposit.
	if session.ReferenceType == "milestone" {
		holdKey := fmt.Sprintf("hold_%d", session.ID)
		_, err := uc.Wallet.PlaceHold(ctx, PlaceHoldInput{
			WalletID:       0,
			AmountMinor:    session.AmountMinor,
			IdempotencyKey: holdKey,
			ReferenceType:  "milestone",
			ReferenceID:    session.ReferenceID,
			Note:           "Escrow hold for milestone",
		})
		if err != nil {
			return session, fmt.Errorf("wallet hold: %w", err)
		}

		// 3. Notify the contract service.
		if err := uc.Contract.UpdateMilestoneStatus(ctx, 0, 0, "FUNDED"); err != nil {
			return session, fmt.Errorf("contract update: %w", err)
		}
	}

	// Mark session completed.
	if err := session.MarkCompleted(now); err != nil {
		return session, err
	}
	if err := uc.Sessions.Update(ctx, session); err != nil {
		return session, fmt.Errorf("update session: %w", err)
	}

	return session, nil
}
