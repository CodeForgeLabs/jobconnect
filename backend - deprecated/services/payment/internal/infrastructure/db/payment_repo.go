package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"jobconnect/payment/internal/application"
	"jobconnect/payment/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentRepo(pool *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{pool: pool}
}

func (r *PaymentRepo) Create(ctx context.Context, s domain.PaymentSession) (domain.PaymentSession, error) {
	query := `
		INSERT INTO payment_sessions (
			user_id, provider, payment_type, status, amount_minor,
			idempotency_key, external_ref, receipt_key, reference_type, reference_id,
			error_message, created_at, updated_at, completed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id
	`
	err := r.pool.QueryRow(ctx, query,
		s.UserID, s.Provider, s.PaymentType, s.Status, s.AmountMinor,
		s.IdempotencyKey, s.ExternalRef, s.ReceiptKey, s.ReferenceType, s.ReferenceID,
		s.ErrorMessage, s.CreatedAt, s.UpdatedAt, s.CompletedAt,
	).Scan(&s.ID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return domain.PaymentSession{}, domain.ErrAlreadyExists
		}
		return domain.PaymentSession{}, fmt.Errorf("insert session: %w", err)
	}
	return s, nil
}

func (r *PaymentRepo) GetByID(ctx context.Context, id int64) (domain.PaymentSession, error) {
	query := `SELECT id, user_id, provider, payment_type, status, amount_minor,
			idempotency_key, external_ref, receipt_key, reference_type, reference_id,
			error_message, created_at, updated_at, completed_at
			FROM payment_sessions WHERE id = $1`

	var s domain.PaymentSession
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.Provider, &s.PaymentType, &s.Status, &s.AmountMinor,
		&s.IdempotencyKey, &s.ExternalRef, &s.ReceiptKey, &s.ReferenceType, &s.ReferenceID,
		&s.ErrorMessage, &s.CreatedAt, &s.UpdatedAt, &s.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PaymentSession{}, domain.ErrNotFound
		}
		return domain.PaymentSession{}, err
	}
	return s, nil
}

func (r *PaymentRepo) GetByIdempotencyKey(ctx context.Context, key string) (domain.PaymentSession, bool, error) {
	query := `SELECT id, user_id, provider, payment_type, status, amount_minor,
			idempotency_key, external_ref, receipt_key, reference_type, reference_id,
			error_message, created_at, updated_at, completed_at
			FROM payment_sessions WHERE idempotency_key = $1`

	var s domain.PaymentSession
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&s.ID, &s.UserID, &s.Provider, &s.PaymentType, &s.Status, &s.AmountMinor,
		&s.IdempotencyKey, &s.ExternalRef, &s.ReceiptKey, &s.ReferenceType, &s.ReferenceID,
		&s.ErrorMessage, &s.CreatedAt, &s.UpdatedAt, &s.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PaymentSession{}, false, nil
		}
		return domain.PaymentSession{}, false, err
	}
	return s, true, nil
}

func (r *PaymentRepo) GetByExternalRef(ctx context.Context, ref string) (domain.PaymentSession, bool, error) {
	query := `SELECT id, user_id, provider, payment_type, status, amount_minor,
			idempotency_key, external_ref, receipt_key, reference_type, reference_id,
			error_message, created_at, updated_at, completed_at
			FROM payment_sessions WHERE external_ref = $1 LIMIT 1`

	var s domain.PaymentSession
	err := r.pool.QueryRow(ctx, query, ref).Scan(
		&s.ID, &s.UserID, &s.Provider, &s.PaymentType, &s.Status, &s.AmountMinor,
		&s.IdempotencyKey, &s.ExternalRef, &s.ReceiptKey, &s.ReferenceType, &s.ReferenceID,
		&s.ErrorMessage, &s.CreatedAt, &s.UpdatedAt, &s.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PaymentSession{}, false, nil
		}
		return domain.PaymentSession{}, false, err
	}
	return s, true, nil
}

func (r *PaymentRepo) Update(ctx context.Context, s domain.PaymentSession) error {
	query := `
		UPDATE payment_sessions SET
			status = $1, external_ref = $2, receipt_key = $3, error_message = $4,
			updated_at = $5, completed_at = $6
		WHERE id = $7
	`
	cmd, err := r.pool.Exec(ctx, query,
		s.Status, s.ExternalRef, s.ReceiptKey, s.ErrorMessage, s.UpdatedAt, s.CompletedAt, s.ID,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PaymentRepo) ListByUserID(ctx context.Context, userID uuid.UUID, filters application.SessionFilters, limit, offset int) ([]domain.PaymentSession, error) {
	query := `SELECT id, user_id, provider, payment_type, status, amount_minor,
			idempotency_key, external_ref, receipt_key, reference_type, reference_id,
			error_message, created_at, updated_at, completed_at
			FROM payment_sessions WHERE user_id = $1`

	args := []any{userID}
	argIdx := 2

	if filters.PaymentType != nil {
		query += fmt.Sprintf(" AND payment_type = $%d", argIdx)
		args = append(args, *filters.PaymentType)
		argIdx++
	}
	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *filters.Status)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []domain.PaymentSession
	for rows.Next() {
		var s domain.PaymentSession
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.Provider, &s.PaymentType, &s.Status, &s.AmountMinor,
			&s.IdempotencyKey, &s.ExternalRef, &s.ReceiptKey, &s.ReferenceType, &s.ReferenceID,
			&s.ErrorMessage, &s.CreatedAt, &s.UpdatedAt, &s.CompletedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}
