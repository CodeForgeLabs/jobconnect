package clients

import (
	"context"
	"fmt"
	"log"

	connectsv1 "jobconnect/api/proto/connects/v1"
	"jobconnect/proposal/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ConnectsClient struct {
	client connectsv1.ConnectsServiceClient
}

func NewConnectsClient(address string) (application.ConnectsClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to connects service at %s", address)
	return &ConnectsClient{
		client: connectsv1.NewConnectsServiceClient(conn),
	}, nil
}

func (c *ConnectsClient) DeductConnects(ctx context.Context, userID uuid.UUID, amount int32, referenceID string) error {
	req := &connectsv1.DeductConnectsRequest{
		UserId:        userID.String(),
		Amount:        amount,
		ReferenceId:   referenceID,
		ReferenceType: "PROPOSAL_SUBMITTED",
	}

	_, err := c.client.DeductConnects(ctx, req)
	if err != nil {
		// e.g. status.Code(err) == codes.FailedPrecondition translates to insufficient funds
		return fmt.Errorf("connects service returned error: %w", err)
	}

	return nil
}
