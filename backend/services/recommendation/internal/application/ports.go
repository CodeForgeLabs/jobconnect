package application

import (
	"context"
	"jobconnect/recommendation/internal/domain"
)

// Ports defining the interfaces for external services that the recommendation service needs.

type JobServiceClient interface {
	GetOpenJobs(ctx context.Context) ([]domain.JobData, error)
	GetJob(ctx context.Context, jobID int64) (domain.JobData, error)
}

type UserServiceClient interface {
	GetFreelancer(ctx context.Context, userID string) (domain.UserData, error)
	GetFreelancers(ctx context.Context) ([]domain.UserData, error)
}
