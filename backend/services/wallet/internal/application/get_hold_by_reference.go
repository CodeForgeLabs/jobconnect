package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/wallet/internal/domain"
)

type GetHoldByReference struct {
	Wallets WalletRepository
}

type GetHoldByReferenceInput struct {
	ReferenceType string
	ReferenceID   string
}

type GetHoldByReferenceOutput struct {
	Hold domain.Hold
}

func (uc *GetHoldByReference) Execute(ctx context.Context, in GetHoldByReferenceInput) (GetHoldByReferenceOutput, error) {
	if uc.Wallets == nil {
		return GetHoldByReferenceOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}
	if strings.TrimSpace(in.ReferenceType) == "" || strings.TrimSpace(in.ReferenceID) == "" {
		return GetHoldByReferenceOutput{}, fmt.Errorf("reference_type and reference_id are required")
	}
	hold, err := uc.Wallets.GetHoldByReference(ctx, strings.TrimSpace(in.ReferenceType), strings.TrimSpace(in.ReferenceID))
	if err != nil {
		return GetHoldByReferenceOutput{}, err
	}
	return GetHoldByReferenceOutput{Hold: hold}, nil
}
