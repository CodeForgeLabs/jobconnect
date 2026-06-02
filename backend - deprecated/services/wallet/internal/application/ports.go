package application

import (
	"context"

	"jobconnect/wallet/internal/domain"

	"github.com/google/uuid"
)

// ==================== REPOSITORY ====================

type WalletRepository interface {
	CreateWallet(ctx context.Context, ownerID uuid.UUID) (domain.WalletAccount, error)
	GetWalletByOwner(ctx context.Context, ownerID uuid.UUID) (domain.WalletAccount, error)
	FetchWalletTransactions(ctx context.Context, walletID int64) ([]domain.WalletTransaction, error)
	CreateDepositTransaction(
		ctx context.Context,
		walletID int64,
		txRef string,
		amountMinor int64,
		description string,
	) (domain.WalletTransaction, error)

	CompleteDeposit(
		ctx context.Context,
		txRef string,
		chapaRef string,
	) error

	GetTransactionByTxRef(
		ctx context.Context,
		txRef string,
	) (domain.WalletTransaction, error)
}

// ==================== INPUT STRUCTS ====================

type CreateWalletInput struct {
	OwnerID uuid.UUID
}

type DepositInput struct {
	WalletID    int64
	AmountMinor int64
	TxRef       string
	Description string
}

type CompleteDepositInput struct {
	TxRef    string
	ChapaRef string
}

// ==================== RESULT STRUCTS ====================

type DepositResult struct {
	Transaction domain.WalletTransaction
}

type WalletResult struct {
	Wallet domain.WalletAccount
}
