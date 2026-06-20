package domain

import "time"

const (
	HoldStatusActive   = "active"
	HoldStatusReleased = "released"
	HoldStatusCaptured = "captured"
)

type Hold struct {
	ID            int64
	WalletID      int64
	ReferenceType string
	ReferenceID   string
	AmountMinor   int64
	CapturedMinor int64
	Status        string
	ExpiresAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (h Hold) RemainingMinor() int64 {
	return h.AmountMinor - h.CapturedMinor
}
