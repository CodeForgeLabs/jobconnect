package application

import (
	"context"
	"fmt"
	"strings"
)

type CreditWalletInternal struct {
	Wallets WalletRepository
}

type CreditWalletInternalInput struct {
	WalletID       int64
	AmountMinor    int64
	IdempotencyKey string
	ReferenceType  string
	ReferenceID    string
	Note           string
}

type CreditWalletInternalOutput struct {
	Result MutationResult
}

func (uc *CreditWalletInternal) Execute(ctx context.Context, in CreditWalletInternalInput) (CreditWalletInternalOutput, error) {
	if uc.Wallets == nil {
		return CreditWalletInternalOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}
	if in.WalletID <= 0 {
		return CreditWalletInternalOutput{}, fmt.Errorf("invalid wallet_id")
	}
	if in.AmountMinor <= 0 {
		return CreditWalletInternalOutput{}, fmt.Errorf("amount_minor must be greater than zero")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return CreditWalletInternalOutput{}, fmt.Errorf("idempotency_key is required")
	}
	result, err := uc.Wallets.CreditInternal(ctx, CreditInput{
		WalletID:       in.WalletID,
		AmountMinor:    in.AmountMinor,
		IdempotencyKey: strings.TrimSpace(in.IdempotencyKey),
		ReferenceType:  strings.TrimSpace(in.ReferenceType),
		ReferenceID:    strings.TrimSpace(in.ReferenceID),
		Note:           strings.TrimSpace(in.Note),
	})
	if err != nil {
		return CreditWalletInternalOutput{}, err
	}
	return CreditWalletInternalOutput{Result: result}, nil
}
