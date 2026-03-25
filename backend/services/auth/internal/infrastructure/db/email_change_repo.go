package db

import (
	"context"
	"time"

	"jobconnect/auth/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailChangeRepo struct {
	pool *pgxpool.Pool
}

func NewEmailChangeRepo(pool *pgxpool.Pool) *EmailChangeRepo {
	return &EmailChangeRepo{pool: pool}
}

func (r *EmailChangeRepo) Upsert(ctx context.Context, userID uuid.UUID, newEmail, otpHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		insert into email_change_requests (user_id, new_email, otp_hash, expires_at)
		values ($1, $2, $3, $4)
		on conflict (user_id) do update
		set new_email = excluded.new_email,
		    otp_hash = excluded.otp_hash,
		    expires_at = excluded.expires_at,
		    attempts = 0,
		    confirmed_at = null,
		    created_at = now()
	`, userID, newEmail, otpHash, expiresAt)
	return err
}

func (r *EmailChangeRepo) Consume(ctx context.Context, userID uuid.UUID, otpPlain string, hasher domain.PasswordHasher, now time.Time) (string, bool, error) {
	var id uuid.UUID
	var newEmail string
	var otpHash string
	var expiresAt time.Time
	var attempts int

	err := r.pool.QueryRow(ctx, `
		select id, new_email, otp_hash, expires_at, attempts
		from email_change_requests
		where user_id = $1 and confirmed_at is null and expires_at > $2
		limit 1
	`, userID, now).Scan(&id, &newEmail, &otpHash, &expiresAt, &attempts)
	if err != nil {
		if isNoRows(err) {
			return "", false, nil
		}
		return "", false, err
	}

	if attempts >= domain.OTPMaxAttempts {
		return "", false, nil
	}

	ok, err := hasher.Verify(otpPlain, otpHash)
	if err != nil {
		return "", false, err
	}
	if !ok {
		_, _ = r.pool.Exec(ctx, `update email_change_requests set attempts = attempts + 1 where id = $1`, id)
		return "", false, nil
	}

	return newEmail, true, nil
}

func (r *EmailChangeRepo) MarkConfirmed(ctx context.Context, userID uuid.UUID, at time.Time) error {
	_, err := r.pool.Exec(ctx, `
		update email_change_requests
		set confirmed_at = $2
		where user_id = $1 and confirmed_at is null
	`, userID, at)
	return err
}
