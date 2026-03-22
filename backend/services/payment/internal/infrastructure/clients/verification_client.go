package clients

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"jobconnect/payment/internal/application"
	verificationv1 "jobconnect/verification/gen/verification/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type VerificationClient struct {
	client verificationv1.VerificationServiceClient
	conn   *grpc.ClientConn
}

func NewVerificationClient(addr string) (*VerificationClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to verification service: %w", err)
	}

	return &VerificationClient{
		client: verificationv1.NewVerificationServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *VerificationClient) Close() error {
	return c.conn.Close()
}

// Ensure VerificationClient implements application.VerificationClient
var _ application.VerificationClient = (*VerificationClient)(nil)

func (c *VerificationClient) IsKYCVerified(ctx context.Context, userID uuid.UUID) (bool, error) {
	req := &verificationv1.GetMyVerificationStatusRequest{
		UserId: userID.String(),
	}

	res, err := c.client.GetMyVerificationStatus(ctx, req)
	if err != nil {
		// If verification service explicitly returns NOT_FOUND, it might mean no record = not verified
		// But let's assume standard error handling for now.
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to get KYC status: %w", err)
	}

	if res.Request == nil {
		return false, nil
	}

	return res.Request.Status == "verified", nil
}
