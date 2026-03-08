package application

import (
	"context"
	"time"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type JobRepository interface {
	Create(ctx context.Context, job domain.Job) (int64, error)
	GetByID(ctx context.Context, jobID int64) (domain.Job, error)
	GetByIDForClient(ctx context.Context, jobID int64, clientID uuid.UUID) (domain.Job, error)
	Update(ctx context.Context, job domain.Job) (domain.Job, error)
	ListByClient(ctx context.Context, clientID uuid.UUID, status string, limit, offset int) ([]domain.Job, error)
	ListOpen(ctx context.Context, limit, offset int) ([]domain.Job, error)
	ListOpenFiltered(ctx context.Context, filter ListOpenFilter) ([]domain.Job, error)
	Close(ctx context.Context, jobID int64, clientID uuid.UUID, reason string, closedAt time.Time) error
}

// ListOpenFilter contains optional filters for the ListOpenJobs query.
type ListOpenFilter struct {
	SearchQuery string
	Skills      []string
	JobType     string
	Limit       int
	Offset      int
}

type Clock interface {
	Now() time.Time
}
