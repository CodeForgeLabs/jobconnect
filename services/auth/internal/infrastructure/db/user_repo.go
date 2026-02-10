package db

import (
	"context"
	"time"

	"jobconnect/auth/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepo implements application.UserRepository.
type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		insert into users (id, email, role, display_name, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6)
		returning id, email, role, display_name, email_verified_at, created_at, updated_at
	`, u.ID, u.Email, u.Role, u.DisplayName, u.CreatedAt, u.UpdatedAt)
	var emailVerifiedAt *time.Time
	err := row.Scan(&u.ID, &u.Email, &u.Role, &u.DisplayName, &emailVerifiedAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return domain.User{}, err
	}
	u.EmailVerifiedAt = emailVerifiedAt
	return u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (domain.User, bool, error) {
	row := r.pool.QueryRow(ctx, `
		select id, email, role, display_name, email_verified_at, created_at, updated_at
		from users where email = $1
	`, email)
	var u domain.User
	var emailVerifiedAt *time.Time
	err := row.Scan(&u.ID, &u.Email, &u.Role, &u.DisplayName, &emailVerifiedAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.User{}, false, nil
		}
		return domain.User{}, false, err
	}
	u.EmailVerifiedAt = emailVerifiedAt
	return u, true, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.User, bool, error) {
	row := r.pool.QueryRow(ctx, `
		select id, email, role, display_name, email_verified_at, created_at, updated_at
		from users where id = $1
	`, id)
	var u domain.User
	var emailVerifiedAt *time.Time
	err := row.Scan(&u.ID, &u.Email, &u.Role, &u.DisplayName, &emailVerifiedAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.User{}, false, nil
		}
		return domain.User{}, false, err
	}
	u.EmailVerifiedAt = emailVerifiedAt
	return u, true, nil
}

func (r *UserRepo) SetEmailVerified(ctx context.Context, userID uuid.UUID, at time.Time) error {
	_, err := r.pool.Exec(ctx, `
		update users set email_verified_at = $2, updated_at = $2 where id = $1
	`, userID, at)
	return err
}
