package application

import (
	"context"
	"time"

	"jobconnect/user/internal/domain"
)

// ProfileRepository persists profiles and role-specific details.
type ProfileRepository interface {
	Create(ctx context.Context, profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) (int64, error)
}

// Clock provides current time (testable).
type Clock interface {
	Now() time.Time
}
