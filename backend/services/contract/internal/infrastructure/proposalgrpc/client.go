package proposalgrpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	shared "jobconnect/events"
	"jobconnect/contract/internal/application"
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
	events *shared.Publisher
}

func NewProposalClient(client proposalv1.ProposalServiceClient, issuer tokenIssuer, events *shared.Publisher) *ProposalClient {
	return &ProposalClient{client: client, issuer: issuer, events: events}
}

func (c *ProposalClient) MarkOfferSent(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string) error {
	if c.events != nil {
		if env, err := shared.NewEnvelope("contract.proposal.mark_offer_sent.requested", fmt.Sprintf("%d", proposalID), "contract-service", 1, map[string]any{"proposal_id": proposalID, "client_id": clientID.String(), "reason": strings.TrimSpace(reason)}, fmt.Sprintf("mark-offer:%d", proposalID), clientID.String()); err == nil {
			_ = c.events.Publish(ctx, env)
		}
	}
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
	ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal", "contract-service")

	_, err = c.client.InternalMarkProposalOfferSent(ctx, &proposalv1.InternalMarkProposalOfferSentRequest{
		ProposalId: proposalID,
		ClientId:   clientID.String(),
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

func (c *ProposalClient) SetHired(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string) error {
	if c.events != nil {
		if env, err := shared.NewEnvelope("contract.proposal.set_hired.requested", fmt.Sprintf("%d", proposalID), "contract-service", 1, map[string]any{"proposal_id": proposalID, "client_id": clientID.String(), "reason": strings.TrimSpace(reason)}, fmt.Sprintf("set-hired:%d", proposalID), clientID.String()); err == nil {
			_ = c.events.Publish(ctx, env)
		}
	}
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
	ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal", "contract-service")

	_, err = c.client.InternalHireProposal(ctx, &proposalv1.InternalHireProposalRequest{
		ProposalId: proposalID,
		ClientId:   clientID.String(),
		RequestId:  hiredRequestID(proposalID, clientID, reason),
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

func (c *ProposalClient) GetProposal(ctx context.Context, proposalID int64, clientID uuid.UUID) (application.ProposalSummary, error) {
	if c.client == nil || c.issuer == nil {
		return application.ProposalSummary{}, fmt.Errorf("proposal client dependencies are not configured")
	}
	if proposalID <= 0 {
		return application.ProposalSummary{}, fmt.Errorf("proposal_id is required")
	}
	if clientID == uuid.Nil {
		return application.ProposalSummary{}, fmt.Errorf("client_id is required")
	}

	token, err := c.issuer.IssueAccessToken(clientID, "client", 2*time.Minute)
	if err != nil {
		return application.ProposalSummary{}, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	res, err := c.client.GetProposal(ctx, &proposalv1.GetProposalRequest{ProposalId: proposalID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound || st.Code() == codes.InvalidArgument || st.Code() == codes.PermissionDenied {
				return application.ProposalSummary{}, err
			}
		}
		return application.ProposalSummary{}, fmt.Errorf("proposal service: %w", err)
	}
	if res.GetProposal() == nil {
		return application.ProposalSummary{}, fmt.Errorf("proposal not found")
	}
	return application.ProposalSummary{
		ID:           res.GetProposal().GetId(),
		JobID:        res.GetProposal().GetJobId(),
		ClientID:     res.GetProposal().GetClientId(),
		FreelancerID: res.GetProposal().GetFreelancerId(),
		Status:       proposalStatus(res.GetProposal().GetStatus()),
	}, nil
}

func (c *ProposalClient) ReleaseOffer(ctx context.Context, proposalID int64, clientID uuid.UUID, reason string) error {
	if c.events != nil {
		if env, err := shared.NewEnvelope("contract.proposal.release_offer.requested", fmt.Sprintf("%d", proposalID), "contract-service", 1, map[string]any{"proposal_id": proposalID, "client_id": clientID.String(), "reason": strings.TrimSpace(reason)}, fmt.Sprintf("release-offer:%d", proposalID), clientID.String()); err == nil {
			_ = c.events.Publish(ctx, env)
		}
	}
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
	ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal", "contract-service")
	_, err = c.client.InternalReleaseHiredProposal(ctx, &proposalv1.InternalReleaseHiredProposalRequest{
		ProposalId: proposalID,
		ClientId:   clientID.String(),
		Reason:     strings.TrimSpace(reason),
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

func proposalStatus(v proposalv1.ProposalStatus) string {
	switch v {
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_SHORTLISTED:
		return "shortlisted"
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_REJECTED:
		return "rejected"
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_OFFER_SENT:
		return "offer_sent"
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_HIRED:
		return "hired"
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_WITHDRAWN:
		return "withdrawn"
	default:
		return "sent"
	}
}

func hiredRequestID(proposalID int64, clientID uuid.UUID, reason string) string {
	normalizedReason := strings.ToLower(strings.TrimSpace(reason))
	if normalizedReason == "" {
		normalizedReason = "unspecified"
	}
	return fmt.Sprintf("contract-offer-%d-%s-%s", proposalID, clientID.String(), normalizedReason)
}
