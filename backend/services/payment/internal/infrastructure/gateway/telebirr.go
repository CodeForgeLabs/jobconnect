package gateway

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/payment/internal/application"
)

// TelebirrGateway is a stub implementation of the Telebirr API.
// A real implementation requires complex RSA/SHA256 signing and Fabric Token generation.
type TelebirrGateway struct {
	appKey string
	appID  string
}

func NewTelebirrGateway(appKey, appID string) *TelebirrGateway {
	return &TelebirrGateway{
		appKey: appKey,
		appID:  appID,
	}
}

var _ application.PaymentGateway = (*TelebirrGateway)(nil)

func (g *TelebirrGateway) InitiateCheckout(ctx context.Context, in application.CheckoutInput) (application.CheckoutResult, error) {
	// MOCK: Return a fake checkout URL
	return application.CheckoutResult{
		CheckoutURL: fmt.Sprintf("https://sandbox.telebirr.et/checkout?tx_ref=%s", in.TxRef),
		ExternalRef: "TB-" + in.TxRef, // Simulate Telebirr's internal reference
	}, nil
}

func (g *TelebirrGateway) VerifyPayment(ctx context.Context, externalRef string) (application.VerifyResult, error) {
	// MOCK: Always return successful verification in sandbox
	// Extract our TxRef from the ExternalRef we created
	txRef := strings.TrimPrefix(externalRef, "TB-")

	return application.VerifyResult{
		Verified:    true,
		AmountMinor: 500000, // Hardcoded for stub (5000 ETB) - in a real app, we'd query Telebirr
		Currency:    "ETB",
		ExternalRef: txRef,
	}, nil
}

func (g *TelebirrGateway) InitiateTransfer(ctx context.Context, in application.TransferInput) (application.TransferResult, error) {
	// MOCK: Simulate a successful payout to a Telebirr Wallet/Phone
	return application.TransferResult{
		TransferID:  "TB-TRANS-" + in.TxRef,
		ExternalRef: in.TxRef,
	}, nil
}

func (g *TelebirrGateway) VerifyWebhookSignature(payload []byte, signature string) error {
	// MOCK: Accept all signatures in sandbox
	return nil
}
