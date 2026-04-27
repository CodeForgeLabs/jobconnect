package db

import (
	"context"
	"errors"
	"fmt"

	"jobconnect/wallet/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepo struct {
	pool *pgxpool.Pool
}

func NewWalletRepo(pool *pgxpool.Pool) *WalletRepo {
	return &WalletRepo{pool: pool}
}

// ==================== WALLET ====================

func (r *WalletRepo) CreateWallet(
	ctx context.Context,
	ownerID uuid.UUID,
) (domain.WalletAccount, error) {

	const q = `
		INSERT INTO wallet_accounts (owner_id, balance_minor)
		VALUES ($1, 0)
		RETURNING id, owner_id, balance_minor, created_at
	`

	var wallet domain.WalletAccount
	err := r.pool.QueryRow(ctx, q, ownerID).Scan(
		&wallet.ID,
		&wallet.OwnerID,
		&wallet.BalanceMinor,
		&wallet.CreatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.WalletAccount{}, fmt.Errorf("wallet already exists")
		}
		return domain.WalletAccount{}, err
	}

	return wallet, nil
}

func (r *WalletRepo) GetWalletByOwner(
	ctx context.Context,
	ownerID uuid.UUID,
) (domain.WalletAccount, error) {

	const q = `
		SELECT id, owner_id, balance_minor, created_at
		FROM wallet_accounts
		WHERE owner_id = $1
	`

	var wallet domain.WalletAccount
	err := r.pool.QueryRow(ctx, q, ownerID).Scan(
		&wallet.ID,
		&wallet.OwnerID,
		&wallet.BalanceMinor,
		&wallet.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WalletAccount{}, fmt.Errorf("wallet not found")
	}

	return wallet, err
}

// ==================== TRANSACTION CREATE ====================

func (r *WalletRepo) CreateDepositTransaction(
	ctx context.Context,
	walletID int64,
	txRef string,
	amountMinor int64,
	description string,
) (domain.WalletTransaction, error) {

	const q = `
		INSERT INTO wallet_transactions (
			wallet_id,
			tx_ref,
			amount_minor,
			tx_type,
			description,
			status
		)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING
			id,
			wallet_id,
			tx_ref,
			chapa_ref,
			amount_minor,
			tx_type,
			description,
			status,
			created_at
	`

	var tx domain.WalletTransaction
	err := r.pool.QueryRow(
		ctx,
		q,
		walletID,
		txRef,
		amountMinor,
		domain.TransactionTypeDeposit,
		description,
	).Scan(
		&tx.ID,
		&tx.WalletID,
		&tx.TxRef,
		&tx.ChapaRef,
		&tx.AmountMinor,
		&tx.TxType,
		&tx.Description,
		&tx.Status,
		&tx.CreatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.WalletTransaction{}, fmt.Errorf("duplicate transaction")
		}
		return domain.WalletTransaction{}, err
	}

	return tx, nil
}

// ==================== VERIFY + COMPLETE ====================

func (r *WalletRepo) CompleteDeposit(
	ctx context.Context,
	txRef string,
	chapaRef string,
) error {

	dbTx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer dbTx.Rollback(ctx)

	var walletID int64
	var amountMinor int64
	var status string

	const getTransaction = `
		SELECT wallet_id, amount_minor, status
		FROM wallet_transactions
		WHERE tx_ref = $1
		FOR UPDATE
	`

	err = dbTx.QueryRow(ctx, getTransaction, txRef).Scan(
		&walletID,
		&amountMinor,
		&status,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("transaction not found")
		}
		return err
	}

	// Prevent double credit
	if status == domain.TransactionSuccess {
		return dbTx.Commit(ctx)
	}

	if status != domain.TransactionPending {
		return fmt.Errorf("invalid transaction state")
	}

	const updateTransaction = `
		UPDATE wallet_transactions
		SET
			status = $2,
			chapa_ref = $3
		WHERE tx_ref = $1
	`

	_, err = dbTx.Exec(
		ctx,
		updateTransaction,
		txRef,
		domain.TransactionSuccess,
		chapaRef,
	)

	if err != nil {
		return err
	}

	const creditWallet = `
		UPDATE wallet_accounts
		SET balance_minor = balance_minor + $2
		WHERE id = $1
	`

	_, err = dbTx.Exec(
		ctx,
		creditWallet,
		walletID,
		amountMinor,
	)

	if err != nil {
		return err
	}

	return dbTx.Commit(ctx)
}

// ==================== TRANSACTION LOOKUP ====================

func (r *WalletRepo) GetTransactionByTxRef(
	ctx context.Context,
	txRef string,
) (domain.WalletTransaction, error) {

	const q = `
		SELECT
			id,
			wallet_id,
			tx_ref,
			chapa_ref,
			amount_minor,
			tx_type,
			description,
			status,
			created_at
		FROM wallet_transactions
		WHERE tx_ref = $1
	`

	var tx domain.WalletTransaction
	err := r.pool.QueryRow(ctx, q, txRef).Scan(
		&tx.ID,
		&tx.WalletID,
		&tx.TxRef,
		&tx.ChapaRef,
		&tx.AmountMinor,
		&tx.TxType,
		&tx.Description,
		&tx.Status,
		&tx.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WalletTransaction{}, fmt.Errorf("transaction not found")
	}

	return tx, err
}
func (r *WalletRepo) FetchWalletTransactions(ctx context.Context, walletID int64) ([]domain.WalletTransaction, error) {

	const q = `
		SELECT
			id,
			wallet_id,
			tx_ref,
			chapa_ref,
			amount_minor,
			tx_type,
			description,
			status,
			created_at
		FROM wallet_transactions
		WHERE wallet_id = $1 AND status = 'success'
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, q, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.WalletTransaction
	for rows.Next() {
		var tx domain.WalletTransaction
		err := rows.Scan(
			&tx.ID,
			&tx.WalletID,
			&tx.TxRef,
			&tx.ChapaRef,
			&tx.AmountMinor,
			&tx.TxType,
			&tx.Description,
			&tx.Status,
			&tx.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// ==================== HELPERS ====================

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
