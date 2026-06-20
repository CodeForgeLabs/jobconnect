package db

import (
	"context"
	"time"

	"jobconnect/auth/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OTPRepo struct {
	pool *pgxpool.Pool
}

func NewOTPRepo(pool *pgxpool.Pool) *OTPRepo {
	return &OTPRepo{pool: pool}
}

func (r *OTPRepo) Create(ctx context.Context, email, purpose, otpHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		insert into otp_codes (email, purpose, otp_hash, expires_at) values ($1, $2, $3, $4)
	`, email, purpose, otpHash, expiresAt)
	return err
}

func (r *OTPRepo) Consume(ctx context.Context, email, purpose, otpPlain string, hasher domain.PasswordHasher) (bool, error) {
	var id, otpHash string
	var expiresAt time.Time
	var attempts int
	row := r.pool.QueryRow(ctx, `
		select id, otp_hash, expires_at, attempts from otp_codes
		where email = $1 and purpose = $2 and expires_at > $3
		order by created_at desc limit 1
	`, email, purpose, time.Now().UTC())
	err := row.Scan(&id, &otpHash, &expiresAt, &attempts)
	if err != nil {
		if isNoRows(err) {
			return false, nil
		}
		return false, err
	}
	if attempts >= domain.OTPMaxAttempts {
		return false, nil
	}
	ok, err := hasher.Verify(otpPlain, otpHash)
	if err != nil {
		return false, err
	}
	if !ok {
		_, _ = r.pool.Exec(ctx, `update otp_codes set attempts = attempts + 1 where id = $1`, id)
		return false, nil
	}
	_, err = r.pool.Exec(ctx, `delete from otp_codes where id = $1`, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *OTPRepo) IncrementAttempts(ctx context.Context, email, purpose string) error {
	_, err := r.pool.Exec(ctx, `
		update otp_codes set attempts = attempts + 1
		where email = $1 and purpose = $2 and expires_at > now()
	`, email, purpose)
	return err
}
