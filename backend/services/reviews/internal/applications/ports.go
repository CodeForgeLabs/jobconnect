package applications

import (
	"context"
	"jobconnect/reviews/internal/domain"
	"time"
)

type ReviewRepository interface {
	Create(ctx context.Context, review domain.Review) (domain.Review, error)

	GetByID(ctx context.Context, id int64) (domain.Review, error)

	Update(ctx context.Context, review domain.Review) (domain.Review, error)

	Delete(ctx context.Context, id int64) error

	ListByUser(ctx context.Context, userID string, role domain.ReviewerRole, limit, offset int) ([]domain.Review, error)
}

type Clock interface {
	Now() time.Time
}
