package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepo struct {
	pool *pgxpool.Pool
}

func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{pool: pool}
}

func (r *SessionRepo) Create(ctx context.Context, userID uuid.UUID, refreshTokenHash string, expiresAt time.Time) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		insert into sessions (user_id, refresh_token_hash, expires_at) values ($1, $2, $3)
		returning id
	`, userID, refreshTokenHash, expiresAt).Scan(&id)
	return id, err
}

func (r *SessionRepo) GetByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (found bool, sessionID, userID uuid.UUID, expiresAt time.Time, revoked bool, err error) {
	var revokedAt *time.Time
	err = r.pool.QueryRow(ctx, `
		select id, user_id, expires_at, revoked_at from sessions
		where refresh_token_hash = $1
	`, refreshTokenHash).Scan(&sessionID, &userID, &expiresAt, &revokedAt)
	if err != nil {
		if isNoRows(err) {
			return false, uuid.Nil, uuid.Nil, time.Time{}, false, nil
		}
		return false, uuid.Nil, uuid.Nil, time.Time{}, false, err
	}
	return true, sessionID, userID, expiresAt, revokedAt != nil, nil
}

func (r *SessionRepo) GetByID(ctx context.Context, sessionID uuid.UUID) (userID uuid.UUID, expiresAt time.Time, revoked bool, err error) {
	var revokedAt *time.Time
	err = r.pool.QueryRow(ctx, `
		select user_id, expires_at, revoked_at from sessions where id = $1
	`, sessionID).Scan(&userID, &expiresAt, &revokedAt)
	if err != nil {
		if isNoRows(err) {
			return uuid.Nil, time.Time{}, false, nil
		}
		return uuid.Nil, time.Time{}, false, err
	}
	return userID, expiresAt, revokedAt != nil, nil
}

func (r *SessionRepo) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `update sessions set revoked_at = $2 where user_id = $1 and revoked_at is null`, userID, now)
	return err
}

func (r *SessionRepo) RevokeByID(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `update sessions set revoked_at = $2 where id = $1`, sessionID, now)
	return err
}

func (r *SessionRepo) UpdateLastUsed(ctx context.Context, sessionID uuid.UUID, at time.Time) error {
	_, err := r.pool.Exec(ctx, `update sessions set last_used_at = $2 where id = $1`, sessionID, at)
	return err
}
