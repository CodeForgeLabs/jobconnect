package application

import (
	"context"
	"fmt"

	"jobconnect/wallet/internal/domain"
)

type GetBalance struct {
	Wallets WalletRepository
}

type GetBalanceInput struct {
	WalletID int64
}

type GetBalanceOutput struct {
	Balance domain.BalanceSnapshot
}

func (uc *GetBalance) Execute(ctx context.Context, in GetBalanceInput) (GetBalanceOutput, error) {
	if uc.Wallets == nil {
		return GetBalanceOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}
	if in.WalletID <= 0 {
		return GetBalanceOutput{}, fmt.Errorf("%w: wallet_id is required", domain.ErrInvalidArgument)
	}
	b, err := uc.Wallets.GetBalance(ctx, in.WalletID)
	if err != nil {
		return GetBalanceOutput{}, err
	}
	return GetBalanceOutput{Balance: b}, nil
}
