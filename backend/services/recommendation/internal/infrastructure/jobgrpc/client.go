package jobgrpc

import (
	"context"

	"google.golang.org/grpc"

	"jobconnect/recommendation/internal/domain"
	jobv1 "jobconnect/job/gen/job/v1"
)

type Client struct {
	grpcClient jobv1.JobServiceClient
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{
		grpcClient: jobv1.NewJobServiceClient(conn),
	}
}

func (c *Client) GetOpenJobs(ctx context.Context) ([]domain.JobData, error) {
	resp, err := c.grpcClient.ListOpenJobs(ctx, &jobv1.ListOpenJobsRequest{
		PageSize: 100, // Reasonable default limit for Phase 1 heuristics
	})
	if err != nil {
		return nil, err
	}

	var jobs []domain.JobData
	for _, j := range resp.Jobs {
		jobs = append(jobs, domain.JobData{
			ID:             j.Id,
			RequiredSkills: j.RequiredSkills,
		})
	}
	return jobs, nil
}

func (c *Client) GetJob(ctx context.Context, jobID int64) (domain.JobData, error) {
	resp, err := c.grpcClient.GetJob(ctx, &jobv1.GetJobRequest{
		JobId: jobID,
	})
	if err != nil {
		return domain.JobData{}, err
	}

	return domain.JobData{
		ID:             resp.Job.Id,
		RequiredSkills: resp.Job.RequiredSkills,
	}, nil
}
