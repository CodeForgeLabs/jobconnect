package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	WalletStatusActive = "active"
	WalletStatusFrozen = "frozen"
)

type WalletAccount struct {
	ID             int64
	OwnerID        uuid.UUID
	Status         string
	AvailableMinor int64
	HeldMinor      int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type BalanceSnapshot struct {
	WalletID       int64
	AvailableMinor int64
	HeldMinor      int64
}

func (b BalanceSnapshot) TotalMinor() int64 {
	return b.AvailableMinor + b.HeldMinor
}

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
