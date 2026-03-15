package application

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type PlaceHold struct {
	Wallets WalletRepository
	Clock   Clock
}

type PlaceHoldCommand struct {
	WalletID       int64
	AmountMinor    int64
	IdempotencyKey string
	ReferenceType  string
	ReferenceID    string
	ExpiresAtUnix  int64
	Note           string
}

type PlaceHoldOutput struct {
	Result HoldMutationResult
}

func (uc *PlaceHold) Execute(ctx context.Context, in PlaceHoldCommand) (PlaceHoldOutput, error) {
	if uc.Wallets == nil {
		return PlaceHoldOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}
	if in.WalletID <= 0 {
		return PlaceHoldOutput{}, fmt.Errorf("invalid wallet_id")
	}
	if in.AmountMinor <= 0 {
		return PlaceHoldOutput{}, fmt.Errorf("amount_minor must be greater than zero")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return PlaceHoldOutput{}, fmt.Errorf("idempotency_key is required")
	}
	if strings.TrimSpace(in.ReferenceType) == "" || strings.TrimSpace(in.ReferenceID) == "" {
		return PlaceHoldOutput{}, fmt.Errorf("reference_type and reference_id are required")
	}
	var expiresAt *time.Time
	if in.ExpiresAtUnix > 0 {
		t := time.Unix(in.ExpiresAtUnix, 0).UTC()
		expiresAt = &t
	}
	result, err := uc.Wallets.PlaceHold(ctx, PlaceHoldInput{
		WalletID:       in.WalletID,
		AmountMinor:    in.AmountMinor,
		IdempotencyKey: strings.TrimSpace(in.IdempotencyKey),
		ReferenceType:  strings.TrimSpace(in.ReferenceType),
		ReferenceID:    strings.TrimSpace(in.ReferenceID),
		ExpiresAt:      expiresAt,
		Note:           strings.TrimSpace(in.Note),
	})
	if err != nil {
		return PlaceHoldOutput{}, err
	}
	return PlaceHoldOutput{Result: result}, nil
}
