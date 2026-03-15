package application

import (
	"context"
	"fmt"
	"strings"
)

type CaptureHold struct {
	Wallets WalletRepository
}

type CaptureHoldCommand struct {
	HoldID             int64
	CaptureAmountMinor int64
	IdempotencyKey     string
	ReferenceType      string
	ReferenceID        string
	Note               string
}

type CaptureHoldOutput struct {
	Result HoldMutationResult
}

func (uc *CaptureHold) Execute(ctx context.Context, in CaptureHoldCommand) (CaptureHoldOutput, error) {
	if uc.Wallets == nil {
		return CaptureHoldOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}
	if in.HoldID <= 0 {
		return CaptureHoldOutput{}, fmt.Errorf("invalid hold_id")
	}
	if in.CaptureAmountMinor <= 0 {
		return CaptureHoldOutput{}, fmt.Errorf("capture_amount_minor must be greater than zero")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return CaptureHoldOutput{}, fmt.Errorf("idempotency_key is required")
	}
	result, err := uc.Wallets.CaptureHold(ctx, CaptureHoldInput{
		HoldID:             in.HoldID,
		CaptureAmountMinor: in.CaptureAmountMinor,
		IdempotencyKey:     strings.TrimSpace(in.IdempotencyKey),
		ReferenceType:      strings.TrimSpace(in.ReferenceType),
		ReferenceID:        strings.TrimSpace(in.ReferenceID),
		Note:               strings.TrimSpace(in.Note),
	})
	if err != nil {
		return CaptureHoldOutput{}, err
	}
	return CaptureHoldOutput{Result: result}, nil
}
