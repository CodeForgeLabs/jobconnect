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
	OwnerID uuid.UUID
}

type GetWalletOutput struct {
	Wallet domain.WalletAccount
}

func (uc *GetWallet) Execute(
	ctx context.Context,
	in GetWalletInput,
) (GetWalletOutput, error) {

	if uc.Wallets == nil {
		return GetWalletOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}

	wallet, err := uc.Wallets.GetWalletByOwner(ctx, in.OwnerID)
	if err != nil {
		return GetWalletOutput{}, err
	}

	return GetWalletOutput{
		Wallet: wallet,
	}, nil
}
