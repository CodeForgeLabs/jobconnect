package domain

import "time"

const (
	LedgerTypeCreditInternal = "credit_internal"
	LedgerTypeDebitInternal  = "debit_internal"
	LedgerTypeHoldPlaced     = "hold_placed"
	LedgerTypeHoldReleased   = "hold_released"
	LedgerTypeHoldCaptured   = "hold_captured"
)

type LedgerEntry struct {
	ID                  int64
	WalletID            int64
	EntryType           string
	AmountMinor         int64
	IdempotencyKey      string
	ReferenceType       string
	ReferenceID         string
	Note                string
	AvailableAfterMinor int64
	HeldAfterMinor      int64
	CreatedAt           time.Time
}
