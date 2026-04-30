package application

import (
	"context"
	"fmt"

	"jobconnect/wallet/internal/domain"
)

type GetTransaction struct {
	Wallets WalletRepository
}

type GetTransactionInput struct {
	TxRef string
}

type GetTransactionOutput struct {
	Transaction domain.WalletTransaction
}

func (uc *GetTransaction) Execute(
	ctx context.Context,
	in GetTransactionInput,
) (GetTransactionOutput, error) {

	if uc.Wallets == nil {
		return GetTransactionOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}

	if in.TxRef == "" {
		return GetTransactionOutput{}, fmt.Errorf("tx_ref is required")
	}

	tx, err := uc.Wallets.GetTransactionByTxRef(
		ctx,
		in.TxRef,
	)

	if err != nil {
		return GetTransactionOutput{}, err
	}

	return GetTransactionOutput{
		Transaction: tx,
	}, nil
}
