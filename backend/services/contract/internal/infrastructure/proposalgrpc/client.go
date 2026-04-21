package proposalgrpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	proposalv1 "jobconnect/proposal/gen/proposal/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type tokenIssuer interface {
	IssueAccessToken(userID uuid.UUID, role string, ttl time.Duration) (string, error)
}

type ProposalClient struct {
	client proposalv1.ProposalServiceClient
	issuer tokenIssuer
}

func NewProposalClient(client proposalv1.ProposalServiceClient, issuer tokenIssuer) *ProposalClient {
	return &ProposalClient{client: client, issuer: issuer}
}

func (c *ProposalClient) SetHired(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string) error {
	if c.client == nil || c.issuer == nil {
		return fmt.Errorf("proposal client dependencies are not configured")
	}
	if proposalID <= 0 {
		return fmt.Errorf("proposal_id is required")
	}
	if clientID == uuid.Nil {
		return fmt.Errorf("client_id is required")
	}

	token, err := c.issuer.IssueAccessToken(clientID, "client", 2*time.Minute)
	if err != nil {
		return err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal", "job-service")

	_, err = c.client.InternalHireProposal(ctx, &proposalv1.InternalHireProposalRequest{
		ProposalId: proposalID,
		ClientId:   clientID.String(),
		RequestId:  fmt.Sprintf("contract-accept-%d-%d", proposalID, time.Now().UnixNano()),
		Note:       strings.TrimSpace(reason),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound || st.Code() == codes.InvalidArgument || st.Code() == codes.PermissionDenied {
				return err
			}
		}
		return fmt.Errorf("proposal service: %w", err)
	}
	return nil
}
