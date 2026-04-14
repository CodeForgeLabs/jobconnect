package clients

import (
	"context"
	"fmt"
	"log"

	proposalv1 "jobconnect/job/gen/proposal/v1"
	"jobconnect/job/internal/application"

	"github.com/google/uuid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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
	req := &proposalv1.ListProposalsByJobRequest{
		JobId:    jobID,
		PageSize: 100, // Handle pagination in real-world scenarios
	}

	forwardCtx := forwardAuthorization(ctx)
	res, err := c.client.ListProposalsByJob(forwardCtx, req)
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
	forwardCtx := forwardAuthorization(ctx)
	res, err := c.client.GetProposal(forwardCtx, &proposalv1.GetProposalRequest{ProposalId: proposalID})
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
	decision, err := applicantStageToDecision(stage)
	if err != nil {
		return err
	}
	forwardCtx := forwardAuthorization(ctx)
	_, err = c.client.SetProposalStatus(forwardCtx, &proposalv1.SetProposalStatusRequest{
		ProposalId: proposalID,
		Decision:   decision,
		Reason:     reason,
	})
	if err != nil {
		return fmt.Errorf("failed to set proposal status: %w", err)
	}
	return nil
}

func (c *ProposalClient) InternalHireProposal(ctx context.Context, proposalID int64, clientID uuid.UUID, requestID string, reason string) error {
	forwardCtx := forwardAuthorization(ctx)
	forwardCtx = metadata.AppendToOutgoingContext(forwardCtx, "x-jobconnect-internal", "job-service")
	_, err := c.client.InternalHireProposal(forwardCtx, &proposalv1.InternalHireProposalRequest{
		ProposalId: proposalID,
		ClientId:   clientID.String(),
		RequestId:  requestID,
		Note:       reason,
	})
	if err != nil {
		return fmt.Errorf("failed to transition proposal to hired: %w", err)
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

func applicantStageToDecision(in string) (proposalv1.ClientDecision, error) {
	switch in {
	case application.ApplicantStageShortlisted:
		return proposalv1.ClientDecision_CLIENT_DECISION_SHORTLISTED, nil
	case application.ApplicantStageRejected:
		return proposalv1.ClientDecision_CLIENT_DECISION_REJECTED, nil
	default:
		return proposalv1.ClientDecision_CLIENT_DECISION_UNSPECIFIED, fmt.Errorf("invalid stage")
	}
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
