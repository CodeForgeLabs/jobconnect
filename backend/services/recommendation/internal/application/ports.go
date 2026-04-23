package application

import (
	"context"
	"jobconnect/recommendation/internal/domain"
)

type JobServiceClient interface {
	ListRecentPublicOpenJobs(ctx context.Context, pageSize int32) ([]domain.JobData, error)
	SearchPublicOpenJobsBySkill(ctx context.Context, skill string, pageSize int32) ([]domain.JobData, error)
	GetJob(ctx context.Context, jobID int64) (domain.JobData, error)
}

type UserServiceClient interface {
	GetFreelancer(ctx context.Context, userID string) (domain.UserData, error)
	GetWorkPreferences(ctx context.Context, userID string) (domain.WorkPreferences, error)
	ListDiscoverableFreelancers(ctx context.Context, skills []string, pageSize int32) ([]domain.FreelancerData, error)
}

type ReviewServiceClient interface {
	GetUserRatingSummary(ctx context.Context, userID string) (domain.RatingSummary, error)
}

type RecommendationCache interface {
	GetRecommendedJobs(userID string) ([]domain.JobRecommendation, bool)
	SetRecommendedJobs(userID string, recommendations []domain.JobRecommendation)
	DeleteRecommendedJobs(userID string) int
	GetRecommendedFreelancers(key string) ([]domain.FreelancerRecommendation, bool)
	SetRecommendedFreelancers(key string, recommendations []domain.FreelancerRecommendation)
	DeleteRecommendedFreelancersForJob(jobID int64) int
	Clear() int
}
