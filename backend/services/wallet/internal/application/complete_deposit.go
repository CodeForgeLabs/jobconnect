package application

import (
	"context"
	"fmt"
)

type CompleteDeposit struct {
	Wallets WalletRepository
}

type CompleteDepositOutput struct {
	Success bool
}

func (uc *CompleteDeposit) Execute(
	ctx context.Context,
	in CompleteDepositInput,
) (CompleteDepositOutput, error) {

	if uc.Wallets == nil {
		return CompleteDepositOutput{}, fmt.Errorf("wallet dependencies are not configured")
	}

	if in.TxRef == "" {
		return CompleteDepositOutput{}, fmt.Errorf("tx_ref is required")
	}

	if in.ChapaRef == "" {
		return CompleteDepositOutput{}, fmt.Errorf("chapa_ref is required")
	}
	// FOR IDEMPOTENCY, CHECK IF THIS TX REF HAS ALREADY BEEN PROCESSED
	// tx, _ := uc.Wallets.GetByTxRef(in.TxRef)

	// if tx.Status == domain.TransactionSuccess {
	// 	return CompleteDepositOutput{Success: true}, nil
	// }
	err := uc.Wallets.CompleteDeposit(
		ctx,
		in.TxRef,
		in.ChapaRef,
	)

	if err != nil {
		return CompleteDepositOutput{}, err
	}

	return CompleteDepositOutput{
		Success: true,
	}, nil
}
