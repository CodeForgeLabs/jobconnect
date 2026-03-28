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
		if p.Status == proposalv1.ProposalStatus_PROPOSAL_STATUS_WITHDRAWN {
			continue
		}
		proposals = append(proposals, application.Proposal{
			ID:            p.Id,
			JobID:         p.JobId,
			ClientID:      p.ClientId,
			FreelancerID:  p.FreelancerId,
			ConnectsSpent: p.ConnectsSpent,
			Status:        proposalStatusToApplicantStage(p.Status),
		})
	}
	return proposals, nil
}

func (c *ProposalClient) GetProposal(ctx context.Context, proposalID int64) (application.Proposal, error) {
	res, err := c.client.GetProposal(ctx, &proposalv1.GetProposalRequest{ProposalId: proposalID})
	if err != nil {
		return application.Proposal{}, fmt.Errorf("failed to get proposal: %w", err)
	}
	if res.Proposal == nil {
		return application.Proposal{}, fmt.Errorf("proposal not found")
	}
	return application.Proposal{
		ID:            res.Proposal.Id,
		JobID:         res.Proposal.JobId,
		ClientID:      res.Proposal.ClientId,
		FreelancerID:  res.Proposal.FreelancerId,
		ConnectsSpent: res.Proposal.ConnectsSpent,
		Status:        proposalStatusToApplicantStage(res.Proposal.Status),
	}, nil
}

func (c *ProposalClient) SetProposalStatus(ctx context.Context, proposalID int64, stage string, reason string) error {
	status, err := applicantStageToProposalStatus(stage)
	if err != nil {
		return err
	}
	_, err = c.client.SetProposalStatus(ctx, &proposalv1.SetProposalStatusRequest{
		ProposalId: proposalID,
		Status:     status,
		Reason:     reason,
	})
	if err != nil {
		return fmt.Errorf("failed to set proposal status: %w", err)
	}
	return nil
}

func proposalStatusToApplicantStage(in proposalv1.ProposalStatus) string {
	switch in {
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_SHORTLISTED:
		return application.ApplicantStageShortlisted
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_REJECTED:
		return application.ApplicantStageRejected
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_HIRED:
		return application.ApplicantStageHired
	default:
		return application.ApplicantStageSent
	}
}

func applicantStageToProposalStatus(in string) (proposalv1.ProposalStatus, error) {
	switch in {
	case application.ApplicantStageShortlisted:
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_SHORTLISTED, nil
	case application.ApplicantStageRejected:
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_REJECTED, nil
	case application.ApplicantStageHired:
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_HIRED, nil
	default:
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED, fmt.Errorf("invalid stage")
	}
}
