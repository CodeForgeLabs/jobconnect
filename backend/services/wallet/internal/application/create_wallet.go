package application

import (
	"context"
	"fmt"

	"jobconnect/wallet/internal/domain"

	"github.com/google/uuid"
)

type CreateWallet struct {
	Wallets WalletRepository
}

type CreateWalletInput struct {
	OwnerID uuid.UUID
}

type CreateWalletOutput struct {
	Wallet domain.WalletAccount
}

func (uc *CreateWallet) Execute(ctx context.Context, in CreateWalletInput) (CreateWalletOutput, error) {
	if uc.Wallets == nil {
		return CreateWalletOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}
	if err := domain.ValidateWalletCreate(in.OwnerID); err != nil {
		return CreateWalletOutput{}, err
	}
	w, err := uc.Wallets.CreateWallet(ctx, in.OwnerID)
	if err != nil {
		return CreateWalletOutput{}, err
	}
	return CreateWalletOutput{Wallet: w}, nil
}
