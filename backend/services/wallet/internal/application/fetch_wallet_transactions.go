package application

import (
	"context"
	"fmt"
	"jobconnect/wallet/internal/domain"
)

type FetchWalletTransactions struct {
	Wallets WalletRepository
}

type FetchWalletTransactionsInput struct {
	WalletID int64
}

type FetchWalletTransactionsOutput struct {
	Transactions []domain.WalletTransaction
}

func (uc *FetchWalletTransactions) Execute(
	ctx context.Context,
	in FetchWalletTransactionsInput,
) (FetchWalletTransactionsOutput, error) {
	if uc.Wallets == nil {
		return FetchWalletTransactionsOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}

	transactions, err := uc.Wallets.FetchWalletTransactions(ctx, in.WalletID)
	if err != nil {
		return FetchWalletTransactionsOutput{}, err
	}

	return FetchWalletTransactionsOutput{
		Transactions: transactions,
	}, nil
}
