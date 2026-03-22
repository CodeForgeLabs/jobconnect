package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ──────────────────────────────────────────────
// Constants
// ──────────────────────────────────────────────

const DefaultCurrency = "ETB"

// Provider constants.
const (
	ProviderChapa    = "chapa"
	ProviderTelebirr = "telebirr"
)

// PaymentType constants.
const (
	TypeDeposit    = "deposit"
	TypeWithdrawal = "withdrawal"
)

// PaymentSession status constants.
const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusRefunded  = "refunded"
)

// SessionTimeout is the maximum age of a PENDING session before it expires.
const SessionTimeout = 30 * time.Minute

// ──────────────────────────────────────────────
// Entity
// ──────────────────────────────────────────────

// PaymentSession tracks a single deposit or withdrawal attempt.
type PaymentSession struct {
	ID              int64
	UserID          uuid.UUID
	Provider        string // "chapa" | "telebirr"
	PaymentType     string // "deposit" | "withdrawal"
	Status          string // "pending" | "completed" | "failed" | "refunded"
	AmountMinor     int64
	Currency        string
	IdempotencyKey  string
	ExternalRef     string // checkout URL or tx reference from gateway
	ReceiptKey      string // MinIO object key for uploaded receipt
	ReferenceType   string // "milestone", "topup", etc.
	ReferenceID     string // e.g. milestone ID
	ErrorMessage    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     *time.Time
}

// ──────────────────────────────────────────────
// State Machine
// ──────────────────────────────────────────────

// validTransitions defines the allowed status transitions.
var validTransitions = map[string][]string{
	StatusPending:   {StatusCompleted, StatusFailed},
	StatusCompleted: {StatusRefunded},
	// StatusFailed and StatusRefunded are terminal.
}

// CanTransition reports whether moving from the current status to newStatus
// is allowed.
func (s *PaymentSession) CanTransition(newStatus string) bool {
	allowed, ok := validTransitions[s.Status]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == newStatus {
			return true
		}
	}
	return false
}

// MarkCompleted transitions the session to COMPLETED.
func (s *PaymentSession) MarkCompleted(now time.Time) error {
	if !s.CanTransition(StatusCompleted) {
		return fmt.Errorf("%w: cannot transition from %q to %q", ErrInvalidTransition, s.Status, StatusCompleted)
	}
	s.Status = StatusCompleted
	s.UpdatedAt = now
	s.CompletedAt = &now
	return nil
}

// MarkFailed transitions the session to FAILED with an error message.
func (s *PaymentSession) MarkFailed(now time.Time, reason string) error {
	if !s.CanTransition(StatusFailed) {
		return fmt.Errorf("%w: cannot transition from %q to %q", ErrInvalidTransition, s.Status, StatusFailed)
	}
	s.Status = StatusFailed
	s.ErrorMessage = reason
	s.UpdatedAt = now
	return nil
}

// MarkRefunded transitions the session to REFUNDED.
func (s *PaymentSession) MarkRefunded(now time.Time) error {
	if !s.CanTransition(StatusRefunded) {
		return fmt.Errorf("%w: cannot transition from %q to %q", ErrInvalidTransition, s.Status, StatusRefunded)
	}
	s.Status = StatusRefunded
	s.UpdatedAt = now
	return nil
}

// IsExpired returns true if the session is PENDING and older than SessionTimeout.
func (s *PaymentSession) IsExpired(now time.Time) bool {
	return s.Status == StatusPending && now.Sub(s.CreatedAt) > SessionTimeout
}

// ──────────────────────────────────────────────
// Validation
// ──────────────────────────────────────────────

// ValidateProvider checks that the provider is one of the supported gateways.
func ValidateProvider(provider string) error {
	switch provider {
	case ProviderChapa, ProviderTelebirr:
		return nil
	default:
		return fmt.Errorf("%w: unsupported provider %q", ErrInvalidArgument, provider)
	}
}

// ValidatePaymentType checks that the payment type is valid.
func ValidatePaymentType(pt string) error {
	switch pt {
	case TypeDeposit, TypeWithdrawal:
		return nil
	default:
		return fmt.Errorf("%w: unsupported payment type %q", ErrInvalidArgument, pt)
	}
}

// ValidateAmountMinor checks that the amount is positive.
func ValidateAmountMinor(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("%w: amount_minor must be positive", ErrInvalidArgument)
	}
	return nil
}

// ValidateDepositInput validates the fields required to initiate a deposit.
func ValidateDepositInput(provider string, amountMinor int64, referenceType, referenceID string) error {
	if err := ValidateProvider(provider); err != nil {
		return err
	}
	if err := ValidateAmountMinor(amountMinor); err != nil {
		return err
	}
	if referenceType == "" {
		return fmt.Errorf("%w: reference_type is required", ErrInvalidArgument)
	}
	if referenceID == "" {
		return fmt.Errorf("%w: reference_id is required", ErrInvalidArgument)
	}
	return nil
}

// ValidateWithdrawalInput validates the fields required to request a withdrawal.
func ValidateWithdrawalInput(provider string, amountMinor int64, bankCode, accountNumber string) error {
	if err := ValidateProvider(provider); err != nil {
		return err
	}
	if err := ValidateAmountMinor(amountMinor); err != nil {
		return err
	}
	if bankCode == "" {
		return fmt.Errorf("%w: bank_code is required", ErrInvalidArgument)
	}
	if accountNumber == "" {
		return fmt.Errorf("%w: account_number is required", ErrInvalidArgument)
	}
	return nil
}
