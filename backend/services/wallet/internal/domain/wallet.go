package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	WalletStatusActive = "active"
	WalletStatusFrozen = "frozen"

	DefaultCurrency = "ETB"
)

type WalletAccount struct {
	ID             int64
	OwnerID        uuid.UUID
	Currency       string
	Status         string
	AvailableMinor int64
	HeldMinor      int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type BalanceSnapshot struct {
	WalletID       int64
	Currency       string
	AvailableMinor int64
	HeldMinor      int64
}

func (b BalanceSnapshot) TotalMinor() int64 {
	return b.AvailableMinor + b.HeldMinor
}

func NormalizeCurrency(currency string) string {
	c := strings.ToUpper(strings.TrimSpace(currency))
	if c == "" {
		return DefaultCurrency
	}
	return c
}

func ValidateWalletCreate(ownerID uuid.UUID, currency string) error {
	if ownerID == uuid.Nil {
		return fmt.Errorf("%w: owner_id is required", ErrInvalidArgument)
	}
	if NormalizeCurrency(currency) == "" {
		return fmt.Errorf("%w: currency is required", ErrInvalidArgument)
	}
	return nil
}

func ValidateAmountMinor(amountMinor int64) error {
	if amountMinor <= 0 {
		return fmt.Errorf("%w: amount_minor must be greater than zero", ErrInvalidArgument)
	}
	return nil
}
