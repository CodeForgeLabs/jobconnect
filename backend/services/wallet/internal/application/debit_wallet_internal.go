package application

import (
	"context"
	"fmt"
	"strings"
)

type DebitWalletInternal struct {
	Wallets WalletRepository
}

type DebitWalletInternalInput struct {
	WalletID       int64
	AmountMinor    int64
	IdempotencyKey string
	ReferenceType  string
	ReferenceID    string
	Note           string
}

type DebitWalletInternalOutput struct {
	Result MutationResult
}

func (uc *DebitWalletInternal) Execute(ctx context.Context, in DebitWalletInternalInput) (DebitWalletInternalOutput, error) {
	if uc.Wallets == nil {
		return DebitWalletInternalOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}
	if in.WalletID <= 0 {
		return DebitWalletInternalOutput{}, fmt.Errorf("invalid wallet_id")
	}
	if in.AmountMinor <= 0 {
		return DebitWalletInternalOutput{}, fmt.Errorf("amount_minor must be greater than zero")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return DebitWalletInternalOutput{}, fmt.Errorf("idempotency_key is required")
	}
	result, err := uc.Wallets.DebitInternal(ctx, DebitInput{
		WalletID:       in.WalletID,
		AmountMinor:    in.AmountMinor,
		IdempotencyKey: strings.TrimSpace(in.IdempotencyKey),
		ReferenceType:  strings.TrimSpace(in.ReferenceType),
		ReferenceID:    strings.TrimSpace(in.ReferenceID),
		Note:           strings.TrimSpace(in.Note),
	})
	if err != nil {
		return DebitWalletInternalOutput{}, err
	}
	return DebitWalletInternalOutput{Result: result}, nil
}
