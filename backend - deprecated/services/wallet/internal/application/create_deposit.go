package application

import (
	"context"
	"fmt"
	"time"

	"jobconnect/wallet/internal/domain"
	"jobconnect/wallet/internal/infrastructure/chapa"
)

type CreateDeposit struct {
	Wallets WalletRepository
	Chapa   *chapa.Client
}

type CreateDepositInput struct {
	WalletID    int64
	AmountMinor int64
	Description string
	Email       string
	Phone       string
}

type CreateDepositOutput struct {
	Transaction domain.WalletTransaction
	PaymentURL  string
}

func (uc *CreateDeposit) Execute(ctx context.Context, in CreateDepositInput) (CreateDepositOutput, error) {

	if uc.Wallets == nil {
		return CreateDepositOutput{}, fmt.Errorf("wallet dependencies not configured")
	}

	if uc.Chapa == nil {
		return CreateDepositOutput{}, fmt.Errorf("chapa client not configured")
	}

	if err := domain.ValidateAmountMinor(in.AmountMinor); err != nil {
		return CreateDepositOutput{}, err
	}
	txRef := fmt.Sprintf("wallet-%d-%d", in.WalletID, time.Now().UnixNano())
	if err := domain.ValidateTxRef(txRef); err != nil {
		return CreateDepositOutput{}, err
	}
	// 2. Call Chapa with dynamic contact info
	url, err := uc.Chapa.InitializePayment(ctx, chapa.PaymentRequest{
		Amount:      float64(in.AmountMinor) / 100,
		TxRef:       txRef,
		Description: in.Description,
		Email:       in.Email,
		PhoneNumber: in.Phone,
		CallbackURL: "https://aorta-contact-scapegoat.ngrok-free.dev/webhook/chapa", // later we will replace it with our actual domain and path
		ReturnURL:   "https://mezgebesibhat.vercel.app/thank-you",                   // later we will replace it with our actual frontend URL

	})

	if err != nil {
		return CreateDepositOutput{}, err
	}
	// 1. Create pending transaction in DB
	tx, err := uc.Wallets.CreateDepositTransaction(
		ctx, in.WalletID, txRef, in.AmountMinor, in.Description,
	)
	if err != nil {
		return CreateDepositOutput{}, err
	}

	return CreateDepositOutput{
		Transaction: tx,
		PaymentURL:  url,
	}, nil
}
