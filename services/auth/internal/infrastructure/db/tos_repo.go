package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TOSRepo struct {
	pool *pgxpool.Pool
}

func NewTOSRepo(pool *pgxpool.Pool) *TOSRepo {
	return &TOSRepo{pool: pool}
}

func (r *TOSRepo) Create(ctx context.Context, userID uuid.UUID, termsVersion, privacyVersion string) error {
	_, err := r.pool.Exec(ctx, `
		insert into tos_acceptances (user_id, terms_version, privacy_version) values ($1, $2, $3)
	`, userID, termsVersion, privacyVersion)
	return err
}
