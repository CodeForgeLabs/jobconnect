package db

import (
	"context"
	"fmt"

	"jobconnect/review/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewRepo struct {
	pool *pgxpool.Pool
}

func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{pool: pool}
}

func (r *ReviewRepo) Create(ctx context.Context, rev domain.Review) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO reviews (contract_id, reviewer_id, reviewee_id, reviewer_role, rating, title, comment, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id
	`, rev.ContractID, rev.ReviewerID, rev.RevieweeID, rev.ReviewerRole,
		rev.Rating, rev.Title, rev.Comment, rev.CreatedAt).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ReviewRepo) GetByID(ctx context.Context, id int64) (domain.Review, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, contract_id, reviewer_id, reviewee_id, reviewer_role, rating, title, comment, created_at, updated_at, reply_comment, replied_at
		FROM reviews
		WHERE id = $1
	`, id)
	rev, err := scanReview(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Review{}, ErrNotFound
		}
		return domain.Review{}, err
	}
	return rev, nil
}

func (r *ReviewRepo) ExistsByContractAndReviewer(ctx context.Context, contractID int64, reviewerID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM reviews WHERE contract_id = $1 AND reviewer_id = $2)
	`, contractID, reviewerID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *ReviewRepo) ListByReviewee(ctx context.Context, revieweeID uuid.UUID, limit, offset int) ([]domain.Review, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, contract_id, reviewer_id, reviewee_id, reviewer_role, rating, title, comment, created_at, updated_at, reply_comment, replied_at
		FROM reviews
		WHERE reviewee_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, revieweeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]domain.Review, 0)
	for rows.Next() {
		rev, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return reviews, nil
}

func (r *ReviewRepo) ListByContract(ctx context.Context, contractID int64) ([]domain.Review, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, contract_id, reviewer_id, reviewee_id, reviewer_role, rating, title, comment, created_at, updated_at, reply_comment, replied_at
		FROM reviews
		WHERE contract_id = $1
		ORDER BY created_at DESC
	`, contractID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]domain.Review, 0)
	for rows.Next() {
		rev, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return reviews, nil
}

func (r *ReviewRepo) GetRatingSummary(ctx context.Context, userID uuid.UUID) (float64, int64, error) {
	var avg *float64
	var count int64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(AVG(rating), 0), COUNT(*)
		FROM reviews
		WHERE reviewee_id = $1
	`, userID).Scan(&avg, &count)
	if err != nil {
		return 0, 0, fmt.Errorf("get rating summary: %w", err)
	}
	if avg == nil {
		return 0, count, nil
	}
	return *avg, count, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanReview(scanner rowScanner) (domain.Review, error) {
	var rev domain.Review
	err := scanner.Scan(
		&rev.ID,
		&rev.ContractID,
		&rev.ReviewerID,
		&rev.RevieweeID,
		&rev.ReviewerRole,
		&rev.Rating,
		&rev.Title,
		&rev.Comment,
		&rev.CreatedAt,
		&rev.UpdatedAt,
		&rev.ReplyComment,
		&rev.RepliedAt,
	)
	if err != nil {
		return domain.Review{}, err
	}
	return rev, nil
}

func (r *ReviewRepo) Update(ctx context.Context, rev domain.Review) (domain.Review, error) {
	res, err := r.pool.Exec(ctx, `
		UPDATE reviews
		SET rating = $2, title = $3, comment = $4, updated_at = $5, reply_comment = $6, replied_at = $7
		WHERE id = $1
	`, rev.ID, rev.Rating, rev.Title, rev.Comment, rev.UpdatedAt, rev.ReplyComment, rev.RepliedAt)
	if err != nil {
		return domain.Review{}, err
	}
	if res.RowsAffected() == 0 {
		return domain.Review{}, ErrNotFound
	}
	return r.GetByID(ctx, rev.ID)
}

func (r *ReviewRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM reviews WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

