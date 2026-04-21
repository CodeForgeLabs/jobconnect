package jobgrpc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	jobv1 "jobconnect/job/gen/job/v1"
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
