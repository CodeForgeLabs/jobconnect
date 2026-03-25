package clients

import (
	"context"
	"fmt"
	"log"

	proposalv1 "jobconnect/api/proto/proposal/v1"
	"jobconnect/job/internal/application"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProposalClient struct {
	client proposalv1.ProposalServiceClient
}

func NewProposalClient(address string) (*ProposalClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to proposal service at %s", address)
	return &ProposalClient{
		client: proposalv1.NewProposalServiceClient(conn),
	}, nil
}

func (c *ProposalClient) ListProposalsByJob(ctx context.Context, jobID int64) ([]application.Proposal, error) {
	// Note: Authentication interceptor or systemic auth bypass might be needed for S2S calls
	// For MVP without strict S2S auth token generation, assuming internal network trust or mock tokens.

	req := &proposalv1.ListProposalsByJobRequest{
		JobId:    jobID,
		PageSize: 100, // Handle pagination in real-world scenarios
	}

	res, err := c.client.ListProposalsByJob(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list proposals: %w", err)
	}

	var proposals []application.Proposal
	for _, p := range res.Proposals {
		if p.Status == proposalv1.ProposalStatus_PROPOSAL_STATUS_SENT || p.Status == proposalv1.ProposalStatus_PROPOSAL_STATUS_SHORTLISTED {
			proposals = append(proposals, application.Proposal{
				ID:            p.Id,
				FreelancerID:  p.FreelancerId,
				ConnectsSpent: p.ConnectsSpent,
			})
		}
	}
	return proposals, nil
}
