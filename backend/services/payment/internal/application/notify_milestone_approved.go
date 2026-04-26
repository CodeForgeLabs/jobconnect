package application

import (
	"context"
	"fmt"

	"jobconnect/payment/internal/domain"
)

// NotifyMilestoneApproved is called by the contract service when a milestone is approved.
// It captures the hold, deducts the platform fee, and credits the freelancer.
type NotifyMilestoneApproved struct {
	Sessions PaymentSessionRepository
	Wallet   WalletClient
	Clock    Clock
}

// NotifyMilestoneApprovedInput is the input from the contract service.
type NotifyMilestoneApprovedInput struct {
	ContractID         int64
	MilestoneID        int64
	HoldID             int64
	FreelancerID       string
	FreelancerWalletID int64
	AmountMinor        int64
}

// PlatformFeePercentage is the platform service fee (e.g., 5%).
const PlatformFeePercentage = 5

func (uc *NotifyMilestoneApproved) Execute(ctx context.Context, in NotifyMilestoneApprovedInput) error {
	now := uc.Clock.Now()

	fee := in.AmountMinor * PlatformFeePercentage / 100
	freelancerAmount := in.AmountMinor - fee

	// 1. Capture the hold — release money from escrow.
	captureKey := fmt.Sprintf("capture_ms_%d", in.MilestoneID)
	if err := uc.Wallet.CaptureHold(ctx, CaptureHoldInput{
		HoldID:             in.HoldID,
		CaptureAmountMinor: freelancerAmount,
		IdempotencyKey:     captureKey,
		ReferenceType:      "milestone",
		ReferenceID:        fmt.Sprintf("%d", in.MilestoneID),
		Note:               fmt.Sprintf("Milestone %d approved — freelancer payout", in.MilestoneID),
	}); err != nil {
		return fmt.Errorf("capture hold: %w", err)
	}

	// 2. Credit the platform wallet with the fee.
	if fee > 0 {
		feeKey := fmt.Sprintf("fee_ms_%d", in.MilestoneID)
		if err := uc.Wallet.CreditWalletInternal(ctx, CreditInput{
			AmountMinor:    fee,
			IdempotencyKey: feeKey,
			ReferenceType:  "platform_fee",
			ReferenceID:    fmt.Sprintf("%d", in.MilestoneID),
			Note:           fmt.Sprintf("Platform fee (%d%%) for milestone %d", PlatformFeePercentage, in.MilestoneID),
		}); err != nil {
			return fmt.Errorf("credit platform fee: %w", err)
		}
	}

	_ = now                    // used for logging/auditing in future
	_ = domain.StatusCompleted // reference to validate import

	return nil
}
