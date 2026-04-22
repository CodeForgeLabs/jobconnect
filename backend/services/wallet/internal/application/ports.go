package application

import (
	"context"
	"time"

	"jobconnect/wallet/internal/domain"

	"github.com/google/uuid"
)

type WalletRepository interface {
	CreateWallet(ctx context.Context, ownerID uuid.UUID) (domain.WalletAccount, error)
	GetWalletByID(ctx context.Context, walletID int64) (domain.WalletAccount, error)
	GetWalletByOwner(ctx context.Context, ownerID uuid.UUID) (domain.WalletAccount, error)
	GetBalance(ctx context.Context, walletID int64) (domain.BalanceSnapshot, error)
	CreditInternal(ctx context.Context, in CreditInput) (MutationResult, error)
	DebitInternal(ctx context.Context, in DebitInput) (MutationResult, error)
	PlaceHold(ctx context.Context, in PlaceHoldInput) (HoldMutationResult, error)
	ReleaseHold(ctx context.Context, in ReleaseHoldInput) (HoldMutationResult, error)
	CaptureHold(ctx context.Context, in CaptureHoldInput) (HoldMutationResult, error)
	ListLedgerEntries(ctx context.Context, walletID int64, limit, offset int) ([]domain.LedgerEntry, error)
}

type Clock interface {
	Now() time.Time
}

type CreditInput struct {
	WalletID       int64
	AmountMinor    int64
	IdempotencyKey string
	ReferenceType  string
	ReferenceID    string
	Note           string
}

type DebitInput struct {
	WalletID       int64
	AmountMinor    int64
	IdempotencyKey string
	ReferenceType  string
	ReferenceID    string
	Note           string
}

type PlaceHoldInput struct {
	WalletID       int64
	AmountMinor    int64
	IdempotencyKey string
	ReferenceType  string
	ReferenceID    string
	ExpiresAt      *time.Time
	Note           string
}

type ReleaseHoldInput struct {
	HoldID         int64
	IdempotencyKey string
	Note           string
}

type CaptureHoldInput struct {
	HoldID             int64
	CaptureAmountMinor int64
	IdempotencyKey     string
	ReferenceType      string
	ReferenceID        string
	Note               string
}

type MutationResult struct {
	Wallet domain.WalletAccount
	Entry  domain.LedgerEntry
}

type HoldMutationResult struct {
	Wallet domain.WalletAccount
	Hold   domain.Hold
	Entry  domain.LedgerEntry
}
