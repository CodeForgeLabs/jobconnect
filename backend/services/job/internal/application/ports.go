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
	ListByClient(ctx context.Context, clientID uuid.UUID, status string, limit, offset int) ([]domain.Job, error)
	ListOpen(ctx context.Context, limit, offset int) ([]domain.Job, error)
	Close(ctx context.Context, jobID int64, clientID uuid.UUID, reason string, closedAt time.Time) error
}

type Clock interface {
	Now() time.Time
}
