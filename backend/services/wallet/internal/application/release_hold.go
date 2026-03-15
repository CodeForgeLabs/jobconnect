package application

import (
	"context"
	"fmt"
	"strings"
)

type ReleaseHold struct {
	Wallets WalletRepository
}

type ReleaseHoldCommand struct {
	HoldID         int64
	IdempotencyKey string
	Note           string
}

type ReleaseHoldOutput struct {
	Result HoldMutationResult
}

func (uc *ReleaseHold) Execute(ctx context.Context, in ReleaseHoldCommand) (ReleaseHoldOutput, error) {
	if uc.Wallets == nil {
		return ReleaseHoldOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}
	if in.HoldID <= 0 {
		return ReleaseHoldOutput{}, fmt.Errorf("invalid hold_id")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return ReleaseHoldOutput{}, fmt.Errorf("idempotency_key is required")
	}
	result, err := uc.Wallets.ReleaseHold(ctx, ReleaseHoldInput{
		HoldID:         in.HoldID,
		IdempotencyKey: strings.TrimSpace(in.IdempotencyKey),
		Note:           strings.TrimSpace(in.Note),
	})
	if err != nil {
		return ReleaseHoldOutput{}, err
	}
	return ReleaseHoldOutput{Result: result}, nil
}
