package jobgrpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	jobv1 "jobconnect/job/gen/job/v1"
	"jobconnect/contract/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type tokenIssuer interface {
	IssueAccessToken(userID uuid.UUID, role string, ttl time.Duration) (string, error)
}

type JobClient struct {
	client jobv1.JobServiceClient
	issuer tokenIssuer
}

func NewJobClient(client jobv1.JobServiceClient, issuer tokenIssuer) *JobClient {
	return &JobClient{client: client, issuer: issuer}
}

func (c *JobClient) GetSummary(ctx context.Context, jobID int64, clientID uuid.UUID) (application.JobSummary, error) {
	if c == nil || c.client == nil || c.issuer == nil {
		return application.JobSummary{}, fmt.Errorf("job client dependencies are not configured")
	}
	if jobID <= 0 {
		return application.JobSummary{}, fmt.Errorf("job_id is required")
	}
	if clientID == uuid.Nil {
		return application.JobSummary{}, fmt.Errorf("client_id is required")
	}

	token, err := c.issuer.IssueAccessToken(clientID, "client", 2*time.Minute)
	if err != nil {
		return application.JobSummary{}, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

	res, err := c.client.GetJobSummary(ctx, &jobv1.GetJobSummaryRequest{JobId: jobID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound || st.Code() == codes.InvalidArgument || st.Code() == codes.PermissionDenied {
				return application.JobSummary{}, err
			}
		}
		return application.JobSummary{}, fmt.Errorf("job service: %w", err)
	}
	summary := res.GetSummary()
	if summary == nil {
		return application.JobSummary{}, fmt.Errorf("job summary not found")
	}
	return application.JobSummary{
		JobID:    summary.GetJobId(),
		ClientID: strings.TrimSpace(summary.GetClientId()),
		Status:   strings.TrimSpace(summary.GetStatus().String()),
		IsOpen:   summary.GetIsOpen(),
		Found:    summary.GetFound(),
	}, nil
}

func (c *JobClient) SetInProgress(ctx context.Context, jobID int64, clientID uuid.UUID) error {
	if c == nil || c.client == nil || c.issuer == nil {
		return fmt.Errorf("job client dependencies are not configured")
	}
	if jobID <= 0 {
		return fmt.Errorf("job_id is required")
	}
	if clientID == uuid.Nil {
		return fmt.Errorf("client_id is required")
	}

	token, err := c.issuer.IssueAccessToken(clientID, "client", 2*time.Minute)
	if err != nil {
		return err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

	_, err = c.client.MarkJobFilled(ctx, &jobv1.MarkJobFilledRequest{JobId: jobID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound || st.Code() == codes.InvalidArgument || st.Code() == codes.PermissionDenied {
				return err
			}
		}
		return fmt.Errorf("job service: %w", err)
	}
	return nil
}
