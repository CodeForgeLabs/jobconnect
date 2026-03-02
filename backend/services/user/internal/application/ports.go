package application

import (
	"context"
	"time"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

// ProfileRepository persists profiles and role-specific details.
type ProfileRepository interface {
	Create(ctx context.Context, profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) (int64, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (domain.Profile, *domain.ClientProfile, *domain.FreelancerProfile, error)
	Update(ctx context.Context, profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) error
	Delete(ctx context.Context, userID uuid.UUID, hardDelete bool, deletedAt time.Time) error
	SaveAvatar(ctx context.Context, avatar domain.Avatar) error
	GetAvatar(ctx context.Context, userID uuid.UUID) (domain.Avatar, error)
	RemoveAvatar(ctx context.Context, userID uuid.UUID) error
}

// AvatarProcessor validates and normalizes avatar images.
type AvatarProcessor interface {
	Process(content []byte, declaredContentType string) (normalized []byte, contentType string, width int, height int, err error)
}

// AvatarModerator decides whether avatar content is safe to keep.
type AvatarModerator interface {
	Moderate(ctx context.Context, content []byte, contentType string) error
}

// Clock provides current time (testable).
type Clock interface {
	Now() time.Time
}
