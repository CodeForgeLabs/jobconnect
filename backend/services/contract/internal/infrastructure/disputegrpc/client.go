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

func (c *Client) GetOpenDisputeID(ctx context.Context, referenceType, referenceID string) (string, error) {
	if c.grpc == nil || c.issuer == nil {
		return "", fmt.Errorf("dispute client dependencies are not configured")
	}
	actorID, _ := uuid.Parse(serviceActorID)
	token, err := c.issuer.IssueAccessToken(actorID, "service", 2*time.Minute)
	if err != nil {
		return "", err
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
		return "", err
	}
	for _, d := range resp.GetDisputes() {
		if d == nil {
			continue
		}
		if d.GetId() > 0 {
			return fmt.Sprintf("%d", d.GetId()), nil
		}
	}
	return "", nil
}

var _ application.DisputeReader = (*Client)(nil)
