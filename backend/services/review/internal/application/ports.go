package application

import (
	"context"
	"time"

	"jobconnect/review/internal/domain"

	"github.com/google/uuid"
)

type ReviewRepository interface {
	Create(ctx context.Context, review domain.Review) (int64, error)
	GetByID(ctx context.Context, id int64) (domain.Review, error)
	ExistsByContractAndReviewer(ctx context.Context, contractID int64, reviewerID uuid.UUID) (bool, error)
	ListByReviewee(ctx context.Context, revieweeID uuid.UUID, limit, offset int) ([]domain.Review, error)
	ListByContract(ctx context.Context, contractID int64) ([]domain.Review, error)
	GetRatingSummary(ctx context.Context, userID uuid.UUID) (avg float64, count int64, err error)
	Update(ctx context.Context, review domain.Review) (domain.Review, error)
	Delete(ctx context.Context, id int64) error
}

type Clock interface {
	Now() time.Time
}
