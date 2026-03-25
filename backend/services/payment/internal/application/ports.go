package application

import (
	"context"
	"time"

	"jobconnect/payment/internal/domain"

	"github.com/google/uuid"
)

// ──────────────────────────────────────────────
// Repository
// ──────────────────────────────────────────────

// PaymentSessionRepository is the persistence port for payment sessions.
type PaymentSessionRepository interface {
	Create(ctx context.Context, session domain.PaymentSession) (domain.PaymentSession, error)
	GetByID(ctx context.Context, id int64) (domain.PaymentSession, error)
	GetByIdempotencyKey(ctx context.Context, key string) (domain.PaymentSession, bool, error)
	GetByExternalRef(ctx context.Context, ref string) (domain.PaymentSession, bool, error)
	Update(ctx context.Context, session domain.PaymentSession) error
	ListByUserID(ctx context.Context, userID uuid.UUID, filters SessionFilters, limit, offset int) ([]domain.PaymentSession, error)
}

// SessionFilters are optional filters for listing sessions.
type SessionFilters struct {
	PaymentType *string
	Status      *string
}

// ──────────────────────────────────────────────
// Payment Gateway (outbound)
// ──────────────────────────────────────────────

// CheckoutResult is returned by InitiateCheckout.
type CheckoutResult struct {
	CheckoutURL string // redirect/deep-link for user
	ExternalRef string // tx reference from provider
}

// TransferResult is returned by InitiateTransfer.
type TransferResult struct {
	TransferID  string
	ExternalRef string
}

// VerifyResult is returned by VerifyPayment.
type VerifyResult struct {
	Verified    bool
	AmountMinor int64
	Currency    string
	ExternalRef string
}

// PaymentGateway abstracts provider-specific payment operations.
type PaymentGateway interface {
	// InitiateCheckout creates a payment session at the provider and returns a checkout URL.
	InitiateCheckout(ctx context.Context, in CheckoutInput) (CheckoutResult, error)

	// VerifyPayment queries the provider to confirm a payment was successful.
	VerifyPayment(ctx context.Context, externalRef string) (VerifyResult, error)

	// InitiateTransfer requests a payout/bank transfer through the provider.
	InitiateTransfer(ctx context.Context, in TransferInput) (TransferResult, error)

	// VerifyWebhookSignature validates the signature of an incoming webhook.
	VerifyWebhookSignature(payload []byte, signature string) error
}

// CheckoutInput contains the data needed to start a checkout session.
type CheckoutInput struct {
	AmountMinor   int64
	Currency      string
	TxRef         string // our idempotency key
	ReturnURL     string
	CallbackURL   string
	Email         string
	FirstName     string
	LastName      string
	PhoneNumber   string
}

// TransferInput contains the data needed for a bank transfer payout.
type TransferInput struct {
	AmountMinor       int64
	Currency          string
	BankCode          string
	AccountNumber     string
	AccountHolderName string
	TxRef             string
}

// ──────────────────────────────────────────────
// Inter-service clients (outbound)
// ──────────────────────────────────────────────

// WalletClient wraps calls to the Wallet Service.
type WalletClient interface {
	CreditWalletInternal(ctx context.Context, in CreditInput) error
	DebitWalletInternal(ctx context.Context, in DebitInput) error
	PlaceHold(ctx context.Context, in PlaceHoldInput) (int64, error) // returns hold ID
	CaptureHold(ctx context.Context, in CaptureHoldInput) error
	GetBalance(ctx context.Context, walletID int64) (BalanceInfo, error)
}

// ContractClient wraps calls to the Contract Service.
type ContractClient interface {
	UpdateMilestoneStatus(ctx context.Context, contractID int64, milestoneID int64, status string) error
}

// VerificationClient wraps calls to the Verification Service.
type VerificationClient interface {
	IsKYCVerified(ctx context.Context, userID uuid.UUID) (bool, error)
}

// ──────────────────────────────────────────────
// Receipt storage (outbound — MinIO/S3)
// ──────────────────────────────────────────────

// ReceiptStore abstracts S3/MinIO operations for payment receipts.
type ReceiptStore interface {
	PutReceipt(ctx context.Context, key string, data []byte, contentType string) error
	GetReceipt(ctx context.Context, key string) ([]byte, error)
	DeleteReceipt(ctx context.Context, key string) error
}

// ──────────────────────────────────────────────
// Clock
// ──────────────────────────────────────────────

// Clock is a time provider that can be faked in tests.
type Clock interface {
	Now() time.Time
}

// ──────────────────────────────────────────────
// Shared DTOs for inter-service calls
// ──────────────────────────────────────────────

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

type BalanceInfo struct {
	AvailableMinor int64
	HeldMinor      int64
	TotalMinor     int64
	Currency       string
}
