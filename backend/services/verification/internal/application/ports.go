package application

import (
	"context"
	"time"

	"jobconnect/verification/internal/domain"

	"github.com/google/uuid"
)

type VerificationRepository interface {
	CreateSubmission(ctx context.Context, req domain.VerificationRequest) (domain.VerificationRequest, error)
	GetLatestByUserID(ctx context.Context, userID uuid.UUID) (domain.VerificationRequest, error)
	GetByID(ctx context.Context, id int64) (domain.VerificationRequest, error)
	ListPending(ctx context.Context, limit, offset int32) ([]domain.VerificationRequest, error)
	Review(ctx context.Context, requestID int64, reviewerID uuid.UUID, decision, rejectionReason, internalNote string, reviewedAt time.Time) (domain.VerificationRequest, error)
	MarkReverificationRequired(ctx context.Context, userID uuid.UUID, reviewerID uuid.UUID, reason string, dueAt time.Time, at time.Time) (domain.VerificationRequest, error)
	AppendEvent(ctx context.Context, event domain.VerificationEvent) error
}

type Clock interface {
	Now() time.Time
}

type VerificationEvidenceObjectStore interface {
	BuildObjectKey(userID uuid.UUID, fileName string) string
	PresignPutObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error)
}
