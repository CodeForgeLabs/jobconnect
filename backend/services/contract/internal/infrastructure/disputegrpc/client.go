package disputegrpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	disputev1 "jobconnect/contract/gen/dispute/v1"
	"jobconnect/contract/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

const serviceActorID = "00000000-0000-0000-0000-00000000c0de"

type tokenIssuer interface {
	IssueAccessToken(userID uuid.UUID, role string, ttl time.Duration) (string, error)
}

type Client struct {
	grpc   disputev1.DisputeServiceClient
	issuer tokenIssuer
}

func NewClient(grpc disputev1.DisputeServiceClient, issuer tokenIssuer) *Client {
	return &Client{grpc: grpc, issuer: issuer}
}

func (c *Client) HasOpenDispute(ctx context.Context, referenceType, referenceID string) (bool, error) {
	if c.grpc == nil || c.issuer == nil {
		return false, fmt.Errorf("dispute client dependencies are not configured")
	}
	actorID, _ := uuid.Parse(serviceActorID)
	token, err := c.issuer.IssueAccessToken(actorID, "service", 2*time.Minute)
	if err != nil {
		return false, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal", "contract-service")

	resp, err := c.grpc.ListDisputes(ctx, &disputev1.ListDisputesRequest{
		ReferenceType: strings.TrimSpace(referenceType),
		ReferenceId:   strings.TrimSpace(referenceID),
		Status:        disputev1.DisputeStatus_DISPUTE_STATUS_OPEN,
		PageSize:      1,
	})
	if err != nil {
		return false, err
	}
	return len(resp.GetDisputes()) > 0, nil
}

var _ application.DisputeReader = (*Client)(nil)
