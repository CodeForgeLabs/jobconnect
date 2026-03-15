package application

import (
	"context"
	"fmt"

	"jobconnect/wallet/internal/domain"

	"github.com/google/uuid"
)

type GetWallet struct {
	Wallets WalletRepository
}

type GetWalletInput struct {
	WalletID int64
	OwnerID  uuid.UUID
	Currency string
}

type GetWalletOutput struct {
	Wallet domain.WalletAccount
}

func (uc *GetWallet) Execute(ctx context.Context, in GetWalletInput) (GetWalletOutput, error) {
	if uc.Wallets == nil {
		return GetWalletOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}
	if in.WalletID > 0 {
		w, err := uc.Wallets.GetWalletByID(ctx, in.WalletID)
		if err != nil {
			return GetWalletOutput{}, err
		}
		return GetWalletOutput{Wallet: w}, nil
	}
	if in.OwnerID == uuid.Nil {
		return GetWalletOutput{}, fmt.Errorf("%w: owner_id is required", domain.ErrInvalidArgument)
	}
	currency := domain.NormalizeCurrency(in.Currency)
	w, err := uc.Wallets.GetWalletByOwnerCurrency(ctx, in.OwnerID, currency)
	if err != nil {
		return GetWalletOutput{}, err
	}
	return GetWalletOutput{Wallet: w}, nil
}
