package clients

import (
	"context"
	"fmt"
	"log"

	connectsv1 "jobconnect/api/proto/connects/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ConnectsClient struct {
	client connectsv1.ConnectsServiceClient
}

func NewConnectsClient(address string) (*ConnectsClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to connects service at %s", address)
	return &ConnectsClient{
		client: connectsv1.NewConnectsServiceClient(conn),
	}, nil
}

func (c *ConnectsClient) RefundConnects(ctx context.Context, userID string, amount int32, referenceID string) error {
	req := &connectsv1.RefundConnectsRequest{
		UserId:        userID,
		Amount:        amount,
		ReferenceId:   referenceID,
		ReferenceType: "JOB_CANCELED_REFUND",
	}

	_, err := c.client.RefundConnects(ctx, req)
	if err != nil {
		return fmt.Errorf("connects service returned error: %w", err)
	}

	return nil
}
