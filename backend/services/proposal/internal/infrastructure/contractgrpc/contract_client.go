package contractgrpc

import (
	"context"
	"fmt"
	"strings"

	contractv1 "jobconnect/contract/gen/contract/v1"
	"jobconnect/proposal/internal/application"
	"jobconnect/proposal/internal/domain"

	"google.golang.org/grpc/metadata"
)

type ContractClient struct {
	client contractv1.ContractServiceClient
}

func NewContractClient(client contractv1.ContractServiceClient) *ContractClient {
	return &ContractClient{client: client}
}

func (c *ContractClient) CreateFromProposal(ctx context.Context, in application.CreateContractFromProposalInput) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("contract client is nil")
	}
	if in.JobID <= 0 {
		return fmt.Errorf("job_id is required")
	}
	if in.ProposalID <= 0 {
		return fmt.Errorf("proposal_id is required")
	}
	contractType := contractv1.ContractType_CONTRACT_TYPE_FIXED
	hourlyRate := 0.0
	fixedTotal := in.BidAmount
	if strings.EqualFold(strings.TrimSpace(in.BidType), domain.BidTypeHourly) {
		contractType = contractv1.ContractType_CONTRACT_TYPE_HOURLY
		hourlyRate = in.BidAmount
		fixedTotal = 0
	}

	forwardCtx := forwardAuthorization(ctx)
	_, err := c.client.CreateContract(forwardCtx, &contractv1.CreateContractRequest{
		FreelancerId:   in.FreelancerID.String(),
		JobId:          in.JobID,
		ProposalId:     in.ProposalID,
		ContractType:   contractType,
		Title:          fmt.Sprintf("Contract for job %d", in.JobID),
		Description:    fmt.Sprintf("Auto-created from proposal %d", in.ProposalID),
		Currency:       "USD",
		HourlyRate:     hourlyRate,
		FixedTotal:     fixedTotal,
		WeeklyHourLimit: 0,
	})
	if err != nil {
		return fmt.Errorf("create contract: %w", err)
	}
	return nil
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
