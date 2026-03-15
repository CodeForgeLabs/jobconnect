package application

import (
	"context"
	"fmt"
	"strconv"

	"jobconnect/wallet/internal/domain"
)

type ListTransactions struct {
	Wallets WalletRepository
}

type ListTransactionsInput struct {
	WalletID  int64
	PageSize  int32
	PageToken string
}

type ListTransactionsOutput struct {
	Transactions  []domain.LedgerEntry
	NextPageToken string
}

func (uc *ListTransactions) Execute(ctx context.Context, in ListTransactionsInput) (ListTransactionsOutput, error) {
	if uc.Wallets == nil {
		return ListTransactionsOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}
	if in.WalletID <= 0 {
		return ListTransactionsOutput{}, fmt.Errorf("%w: wallet_id is required", domain.ErrInvalidArgument)
	}
	limit := int(in.PageSize)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := 0
	if in.PageToken != "" {
		n, err := strconv.Atoi(in.PageToken)
		if err != nil || n < 0 {
			return ListTransactionsOutput{}, fmt.Errorf("%w: invalid page_token", domain.ErrInvalidArgument)
		}
		offset = n
	}
	items, err := uc.Wallets.ListLedgerEntries(ctx, in.WalletID, limit+1, offset)
	if err != nil {
		return ListTransactionsOutput{}, err
	}
	next := ""
	if len(items) > limit {
		next = strconv.Itoa(offset + limit)
		items = items[:limit]
	}
	return ListTransactionsOutput{Transactions: items, NextPageToken: next}, nil
}
