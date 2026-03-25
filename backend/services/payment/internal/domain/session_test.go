package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name   string
		from   string
		to     string
		expect bool
	}{
		{"pending->completed", StatusPending, StatusCompleted, true},
		{"pending->failed", StatusPending, StatusFailed, true},
		{"pending->refunded", StatusPending, StatusRefunded, false},
		{"completed->refunded", StatusCompleted, StatusRefunded, true},
		{"completed->failed", StatusCompleted, StatusFailed, false},
		{"failed->completed", StatusFailed, StatusCompleted, false},
		{"refunded->anything", StatusRefunded, StatusCompleted, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PaymentSession{Status: tt.from}
			got := s.CanTransition(tt.to)
			if got != tt.expect {
				t.Errorf("CanTransition(%q→%q) = %v, want %v", tt.from, tt.to, got, tt.expect)
			}
		})
	}
}

func TestMarkCompleted(t *testing.T) {
	now := time.Now()
	s := &PaymentSession{Status: StatusPending, CreatedAt: now}

	if err := s.MarkCompleted(now); err != nil {
		t.Fatalf("MarkCompleted: unexpected error: %v", err)
	}
	if s.Status != StatusCompleted {
		t.Errorf("status = %q, want %q", s.Status, StatusCompleted)
	}
	if s.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}

	// Cannot complete again.
	if err := s.MarkCompleted(now); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestMarkFailed(t *testing.T) {
	now := time.Now()
	s := &PaymentSession{Status: StatusPending, CreatedAt: now}

	if err := s.MarkFailed(now, "timeout"); err != nil {
		t.Fatalf("MarkFailed: unexpected error: %v", err)
	}
	if s.Status != StatusFailed {
		t.Errorf("status = %q, want %q", s.Status, StatusFailed)
	}
	if s.ErrorMessage != "timeout" {
		t.Errorf("error_message = %q, want %q", s.ErrorMessage, "timeout")
	}
}

func TestMarkRefunded(t *testing.T) {
	now := time.Now()
	s := &PaymentSession{Status: StatusCompleted, CreatedAt: now}

	if err := s.MarkRefunded(now); err != nil {
		t.Fatalf("MarkRefunded: unexpected error: %v", err)
	}
	if s.Status != StatusRefunded {
		t.Errorf("status = %q, want %q", s.Status, StatusRefunded)
	}

	// Cannot refund from failed.
	s2 := &PaymentSession{Status: StatusFailed}
	if err := s2.MarkRefunded(now); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestIsExpired(t *testing.T) {
	now := time.Now()

	fresh := &PaymentSession{Status: StatusPending, CreatedAt: now.Add(-5 * time.Minute)}
	if fresh.IsExpired(now) {
		t.Error("5-minute old session should not be expired")
	}

	old := &PaymentSession{Status: StatusPending, CreatedAt: now.Add(-31 * time.Minute)}
	if !old.IsExpired(now) {
		t.Error("31-minute old session should be expired")
	}

	// Completed sessions never expire.
	completed := &PaymentSession{Status: StatusCompleted, CreatedAt: now.Add(-1 * time.Hour)}
	if completed.IsExpired(now) {
		t.Error("completed session should not be expired")
	}
}

func TestValidateProvider(t *testing.T) {
	if err := ValidateProvider(ProviderChapa); err != nil {
		t.Errorf("chapa should be valid: %v", err)
	}
	if err := ValidateProvider(ProviderTelebirr); err != nil {
		t.Errorf("telebirr should be valid: %v", err)
	}
	if err := ValidateProvider("cbe"); err == nil {
		t.Error("cbe should be invalid")
	}
}

func TestValidateDepositInput(t *testing.T) {
	if err := ValidateDepositInput(ProviderChapa, 5000, "milestone", "123"); err != nil {
		t.Errorf("valid deposit should pass: %v", err)
	}
	if err := ValidateDepositInput("invalid", 5000, "milestone", "123"); err == nil {
		t.Error("invalid provider should fail")
	}
	if err := ValidateDepositInput(ProviderChapa, 0, "milestone", "123"); err == nil {
		t.Error("zero amount should fail")
	}
	if err := ValidateDepositInput(ProviderChapa, 5000, "", "123"); err == nil {
		t.Error("empty reference_type should fail")
	}
}

func TestValidateWithdrawalInput(t *testing.T) {
	if err := ValidateWithdrawalInput(ProviderChapa, 5000, "CBE", "1000123456"); err != nil {
		t.Errorf("valid withdrawal should pass: %v", err)
	}
	if err := ValidateWithdrawalInput(ProviderChapa, 5000, "", "1000123456"); err == nil {
		t.Error("empty bank_code should fail")
	}
	if err := ValidateWithdrawalInput(ProviderChapa, 5000, "CBE", ""); err == nil {
		t.Error("empty account_number should fail")
	}
}

// Suppress unused import warning.
var _ = uuid.Nil
