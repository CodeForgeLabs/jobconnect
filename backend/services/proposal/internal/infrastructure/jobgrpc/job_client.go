package jobgrpc

import (
	"context"
	"fmt"

	jobv1 "jobconnect/job/gen/job/v1"
	"jobconnect/proposal/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type JobClient struct {
	client jobv1.JobServiceClient
}

func NewJobClient(client jobv1.JobServiceClient) *JobClient {
	return &JobClient{client: client}
}

func (c *JobClient) GetJobSummary(ctx context.Context, jobID int64) (application.JobSummary, error) {
	if c == nil || c.client == nil {
		return application.JobSummary{}, fmt.Errorf("job client is nil")
	}
	if jobID <= 0 {
		return application.JobSummary{}, fmt.Errorf("job_id is required")
	}

	forwardCtx := forwardAuthorization(ctx)
	resp, err := c.client.GetJobSummary(forwardCtx, &jobv1.GetJobSummaryRequest{JobId: jobID})
	if err != nil {
		return application.JobSummary{}, fmt.Errorf("get job summary: %w", err)
	}
	if resp.GetSummary() == nil || !resp.GetSummary().GetFound() {
		return application.JobSummary{JobID: jobID, Found: false}, nil
	}

	clientID, err := uuid.Parse(resp.GetSummary().GetClientId())
	if err != nil {
		return application.JobSummary{}, fmt.Errorf("invalid client_id from job service")
	}

	return application.JobSummary{
		JobID:    resp.GetSummary().GetJobId(),
		ClientID: clientID,
		Status:   resp.GetSummary().GetStatus().String(),
		IsOpen:   resp.GetSummary().GetIsOpen(),
		Found:    resp.GetSummary().GetFound(),
	}, nil
}

func (c *JobClient) MarkJobFilled(ctx context.Context, jobID int64) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("job client is nil")
	}
	if jobID <= 0 {
		return fmt.Errorf("job_id is required")
	}
	forwardCtx := forwardAuthorization(ctx)
	_, err := c.client.MarkJobFilled(forwardCtx, &jobv1.MarkJobFilledRequest{JobId: jobID})
	if err != nil {
		return fmt.Errorf("mark job filled: %w", err)
	}
	return nil
}

func forwardAuthorization(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", vals[0]))
}
