package clients

import (
	"context"
	"fmt"

	contractv1 "jobconnect/contract/gen/contract/v1"
	"jobconnect/payment/internal/application"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ContractClient struct {
	client contractv1.ContractServiceClient
	conn   *grpc.ClientConn
}

func NewContractClient(addr string) (*ContractClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to contract service: %w", err)
	}

	return &ContractClient{
		client: contractv1.NewContractServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *ContractClient) Close() error {
	return c.conn.Close()
}

// Ensure ContractClient implements application.ContractClient
var _ application.ContractClient = (*ContractClient)(nil)

func (c *ContractClient) UpdateMilestoneStatus(ctx context.Context, contractID int64, milestoneID int64, status string) error {
	var protoStatus contractv1.MilestoneStatus
	switch status {
	case "FUNDED":
		protoStatus = contractv1.MilestoneStatus_MILESTONE_STATUS_FUNDED
	case "APPROVED":
		protoStatus = contractv1.MilestoneStatus_MILESTONE_STATUS_APPROVED
	// Add other mappings if needed, but Payment mainly cares about FUNDED
	default:
		return fmt.Errorf("unsupported milestone status to update: %s", status)
	}

	req := &contractv1.UpdateMilestoneStatusRequest{
		ContractId:  contractID,
		MilestoneId: milestoneID,
		Status:      protoStatus,
	}

	_, err := c.client.UpdateMilestoneStatus(ctx, req)
	if err != nil {
		// Just a generic error wrap
		return fmt.Errorf("failed to update milestone status: %w", err)
	}

	return nil
}
