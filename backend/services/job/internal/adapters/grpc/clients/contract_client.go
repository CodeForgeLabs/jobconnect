package clients

import (
	"context"
	"fmt"
	"log"
	"strings"

	contractv1 "jobconnect/job/gen/contract/v1"
	"jobconnect/job/internal/application"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ContractClient struct {
	client contractv1.ContractServiceClient
}

func NewContractClient(address string) (*ContractClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to contract service at %s", address)
	return &ContractClient{client: contractv1.NewContractServiceClient(conn)}, nil
}

func (c *ContractClient) CreateFromProposal(ctx context.Context, in application.CreateContractFromProposalInput) (int64, error) {
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
	if strings.EqualFold(strings.TrimSpace(in.BidType), "hourly") {
		contractType = contractv1.ContractType_CONTRACT_TYPE_HOURLY
		hourlyRate = in.BidAmount
		fixedTotal = 0
	}

	forwardCtx := forwardAuthorization(ctx)
	res, err := c.client.CreateContract(forwardCtx, &contractv1.CreateContractRequest{
		FreelancerId:    in.FreelancerID,
		JobId:           in.JobID,
		ProposalId:      in.ProposalID,
		ContractType:    contractType,
		Title:           fmt.Sprintf("Contract for job %d", in.JobID),
		Description:     fmt.Sprintf("Auto-created from proposal %d", in.ProposalID),
		HourlyRate:      hourlyRate,
		FixedTotal:      fixedTotal,
		WeeklyHourLimit: 0,
	})
	if err != nil {
		return fmt.Errorf("create contract: %w", err)
	}
	if res == nil || res.Contract == nil {
		return 0, fmt.Errorf("create contract: empty response")
	}
	return res.Contract.Id, nil
}
