package application

import (
	"context"
	"jobconnect/recommendation/internal/domain"
)

type JobServiceClient interface {
	ListRecentPublicOpenJobs(ctx context.Context, pageSize int32) ([]domain.JobData, error)
	SearchPublicOpenJobsBySkill(ctx context.Context, skill string, pageSize int32) ([]domain.JobData, error)
}

type UserServiceClient interface {
	GetFreelancer(ctx context.Context, userID string) (domain.UserData, error)
	GetWorkPreferences(ctx context.Context, userID string) (domain.WorkPreferences, error)
}

type RecommendationCache interface {
	GetRecommendedJobs(userID string) ([]domain.JobRecommendation, bool)
	SetRecommendedJobs(userID string, recommendations []domain.JobRecommendation)
}
