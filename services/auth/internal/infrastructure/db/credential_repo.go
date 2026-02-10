package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CredentialRepo struct {
	pool *pgxpool.Pool
}

func NewCredentialRepo(pool *pgxpool.Pool) *CredentialRepo {
	return &CredentialRepo{pool: pool}
}

func (r *CredentialRepo) Create(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	_, err := r.pool.Exec(ctx, `
		insert into credentials (user_id, password_hash) values ($1, $2)
	`, userID, passwordHash)
	return err
}

func (r *CredentialRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (passwordHash string, found bool, err error) {
	err = r.pool.QueryRow(ctx, `select password_hash from credentials where user_id = $1`, userID).Scan(&passwordHash)
	if err != nil {
		if isNoRows(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return passwordHash, true, nil
}
