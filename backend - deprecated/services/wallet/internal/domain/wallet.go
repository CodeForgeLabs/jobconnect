package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	WalletStatusActive = "active"

	TransactionPending = "pending"
	TransactionSuccess = "success"
	TransactionFailed  = "failed"
)

// ==================== WALLET ====================

type WalletAccount struct {
	ID           int64
	OwnerID      uuid.UUID
	BalanceMinor int64
	CreatedAt    time.Time
}

type BalanceSnapshot struct {
	WalletID     int64
	BalanceMinor int64
}

// ==================== TRANSACTIONS ====================

type WalletTransaction struct {
	ID          int64
	WalletID    int64
	TxRef       string
	ChapaRef    *string
	AmountMinor int64
	TxType      string
	Description string
	Status      string
	CreatedAt   time.Time
}

const (
	TransactionTypeDeposit    = "deposit"
	TransactionTypeWithdrawal = "withdrawal"
	TransactionTypePayment    = "payment"
)

// ==================== VALIDATION ====================

func ValidateWalletCreate(ownerID uuid.UUID) error {
	if ownerID == uuid.Nil {
		return fmt.Errorf("%w: owner_id is required", ErrInvalidArgument)
	}
	return nil
}

func ValidateAmountMinor(amountMinor int64) error {
	if amountMinor <= 0 {
		return fmt.Errorf("%w: amount_minor must be greater than zero", ErrInvalidArgument)
	}
	return nil
}

func ValidateTxRef(txRef string) error {
	if txRef == "" {
		return fmt.Errorf("%w: tx_ref is required", ErrInvalidArgument)
	}
	return nil
}
