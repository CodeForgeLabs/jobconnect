package db

import (
	"context"
	"errors"
	"fmt"

	"jobconnect/services/connects/internal/application"
	"jobconnect/services/connects/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConnectsRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresConnectsRepository(pool *pgxpool.Pool) application.ConnectsRepository {
	return &PostgresConnectsRepository{pool: pool}
}

func (r *PostgresConnectsRepository) GetBalance(ctx context.Context, userID string) (domain.ConnectsBalance, error) {
	var b domain.ConnectsBalance
	err := r.pool.QueryRow(ctx, "SELECT user_id, balance, version, updated_at FROM connects_balances WHERE user_id = $1", userID).
		Scan(&b.UserID, &b.Balance, &b.Version, &b.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// If no row exists, we treat it as a 0 balance row with version 0
			// The transaction execution will handle INSERTing the row.
			return domain.ConnectsBalance{UserID: userID, Balance: 0, Version: 0}, nil
		}
		return domain.ConnectsBalance{}, fmt.Errorf("failed to get balance: %w", err)
	}

	return b, nil
}

func (r *PostgresConnectsRepository) ExecuteTransaction(ctx context.Context, tx domain.ConnectsTransaction) (domain.ConnectsBalance, error) {
	dbTx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.ConnectsBalance{}, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer dbTx.Rollback(ctx)

	// 1. Get current balance & version
	var b domain.ConnectsBalance
	err = dbTx.QueryRow(ctx, "SELECT balance, version FROM connects_balances WHERE user_id = $1", tx.UserID).
		Scan(&b.Balance, &b.Version)
	
	rowExists := true
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			rowExists = false
			b.Balance = 0
			b.Version = 0
		} else {
			return domain.ConnectsBalance{}, fmt.Errorf("query balance: %w", err)
		}
	}

	// 2. Calculate new balance
	newBalance := b.Balance + tx.Amount
	if newBalance < 0 {
		return domain.ConnectsBalance{}, domain.ErrInsufficientBalance
	}
	newVersion := b.Version + 1

	// 3. Update or Insert balance explicitly using the exact version to ensure no concurrent modification
	var affected int64
	if rowExists {
		res, err := dbTx.Exec(ctx,
			"UPDATE connects_balances SET balance = $1, version = $2, updated_at = NOW() WHERE user_id = $3 AND version = $4",
			newBalance, newVersion, tx.UserID, b.Version)
		if err != nil {
			return domain.ConnectsBalance{}, fmt.Errorf("update balance: %w", err)
		}
		affected = res.RowsAffected()
	} else {
		res, err := dbTx.Exec(ctx,
			"INSERT INTO connects_balances (user_id, balance, version) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
			tx.UserID, newBalance, newVersion)
		if err != nil {
			return domain.ConnectsBalance{}, fmt.Errorf("insert balance: %w", err)
		}
		affected = res.RowsAffected()
	}

	if affected == 0 {
		return domain.ConnectsBalance{}, fmt.Errorf("concurrent modification error or unhandled row existence case")
	}

	// 4. Insert Transaction Log
	_, err = dbTx.Exec(ctx,
		"INSERT INTO connects_transactions (user_id, amount, type, reference_id, reference_type) VALUES ($1, $2, $3, $4, $5)",
		tx.UserID, tx.Amount, tx.Type, tx.ReferenceID, tx.ReferenceType)
	if err != nil {
		// e.g. violates unique idx_connects_transactions_idempotency
		// we should probably check if it's that specific error in production, but returning error will rollback 
		return domain.ConnectsBalance{}, fmt.Errorf("insert transaction: %w", err)
	}

	err = dbTx.Commit(ctx)
	if err != nil {
		return domain.ConnectsBalance{}, fmt.Errorf("commit tx: %w", err)
	}

	return domain.ConnectsBalance{
		UserID:  tx.UserID,
		Balance: newBalance,
		Version: newVersion,
	}, nil
}
