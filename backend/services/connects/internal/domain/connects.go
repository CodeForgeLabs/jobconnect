package domain

import (
	"errors"
	"time"
)

var (
	ErrInsufficientBalance    = errors.New("insufficient connects balance")
	ErrTransactionExists      = errors.New("transaction with this reference already exists")
	ErrNegativeDeduction      = errors.New("deduction amount must be positive")
	ErrInvalidConnectsMinimum = errors.New("minimum 1 connect required for this action")
)

type ConnectsBalance struct {
	UserID    string
	Balance   int32
	Version   int32
	UpdatedAt time.Time
}

type TransactionType string

const (
	TxTypeProposalSubmitted TransactionType = "PROPOSAL_SUBMITTED"
	TxTypeJobCanceledRefund TransactionType = "JOB_CANCELED_REFUND"
	TxTypeRegisterBonus     TransactionType = "REGISTER_BONUS"
	TxTypeFiatPurchase      TransactionType = "FIAT_PURCHASE"
)

type ConnectsTransaction struct {
	ID            int64
	UserID        string
	Amount        int32
	Type          TransactionType
	ReferenceID   string
	ReferenceType string
	CreatedAt     time.Time
}
