package clients

import (
	"context"
	"fmt"
	"log"

	connectsv1 "jobconnect/api/proto/connects/v1"

	"github.com/google/uuid"
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

func (c *ConnectsClient) GrantInitialConnects(ctx context.Context, userID uuid.UUID) error {
	req := &connectsv1.GrantInitialConnectsRequest{
		UserId: userID.String(),
	}

	_, err := c.client.GrantInitialConnects(ctx, req)
	if err != nil {
		return fmt.Errorf("connects service returned error: %w", err)
	}

	return nil
}
