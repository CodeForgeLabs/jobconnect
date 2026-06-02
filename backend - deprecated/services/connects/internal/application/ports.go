package application

import (
	"context"
	"jobconnect/services/connects/internal/domain"
)

type ConnectsRepository interface {
	// GetBalance retrieves the user's current balance and version for optimistic locking.
	GetBalance(ctx context.Context, userID string) (domain.ConnectsBalance, error)

	// ExecuteTransaction applies a deduction or credit to the balance and records the transaction.
	// It uses optimistic locking under the hood to prevent negative balances.
	ExecuteTransaction(ctx context.Context, tx domain.ConnectsTransaction) (domain.ConnectsBalance, error)
}
