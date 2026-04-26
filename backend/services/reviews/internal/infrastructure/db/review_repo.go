package db

import (
	"context"
	"jobconnect/reviews/internal/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewRepo struct {
	pool *pgxpool.Pool
}

func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{pool: pool}
}

func (r *ReviewRepo) Create(ctx context.Context, review domain.Review) (domain.Review, error) {

	query := `
		INSERT INTO reviews (
			contract_id, client_id, freelancer_id,
			reviewer_role, rating, title, comment,
			created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		review.ContractID,
		review.ClientID,
		review.FreelancerID,
		review.ReviewerRole,
		review.Rating,
		review.Title,
		review.Comment,
		review.CreatedAt,
		review.UpdatedAt,
	).Scan(&review.ID)

	if err != nil {
		println(err.Error())
		return domain.Review{}, err
	}

	return review, nil
}

func (r *ReviewRepo) GetByID(ctx context.Context, id int64) (domain.Review, error) {
	var review domain.Review

	query := `
		SELECT id, contract_id, client_id, freelancer_id,
		       reviewer_role, rating, title, comment,
		       created_at, updated_at
		FROM reviews
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&review.ID,
		&review.ContractID,
		&review.ClientID,
		&review.FreelancerID,
		&review.ReviewerRole,
		&review.Rating,
		&review.Title,
		&review.Comment,
		&review.CreatedAt,
		&review.UpdatedAt,
	)

	return review, err
}
func (r *ReviewRepo) Update(ctx context.Context, review domain.Review) (domain.Review, error) {
	query := `
		UPDATE reviews
		SET rating = $1,
		    title = $2,
		    comment = $3,
		    updated_at = $4
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		review.Rating,
		review.Title,
		review.Comment,
		time.Now(),
		review.ID,
	).Scan(&review.UpdatedAt)

	return review, err
}

func (r *ReviewRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM reviews WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *ReviewRepo) ListByUser(
	ctx context.Context,
	userID string,
	role domain.ReviewerRole,
	limit, offset int,
) ([]domain.Review, error) {

	query := `
		SELECT id, contract_id, client_id, freelancer_id,
		       reviewer_role, rating, title, comment,
		       created_at, updated_at
		FROM reviews
		WHERE (client_id = $1 OR freelancer_id = $1)
		  AND reviewer_role = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, userID, role, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []domain.Review

	for rows.Next() {
		var r2 domain.Review

		err := rows.Scan(
			&r2.ID,
			&r2.ContractID,
			&r2.ClientID,
			&r2.FreelancerID,
			&r2.ReviewerRole,
			&r2.Rating,
			&r2.Title,
			&r2.Comment,
			&r2.CreatedAt,
			&r2.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		reviews = append(reviews, r2)
	}

	return reviews, nil
}
