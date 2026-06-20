package clients

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	contractv1 "jobconnect/job/gen/contract/v1"
	"jobconnect/job/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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

func (c *ContractClient) GetJobOfferState(ctx context.Context, jobID int64, clientID uuid.UUID) (application.ContractState, error) {
	if c == nil || c.client == nil {
		return application.ContractState{}, fmt.Errorf("contract client is nil")
	}
	if jobID <= 0 {
		return application.ContractState{}, fmt.Errorf("job_id is required")
	}
	if clientID == uuid.Nil {
		return application.ContractState{}, fmt.Errorf("client_id is required")
	}
	forwardCtx := forwardAuthorization(ctx)
	forwardCtx = metadata.AppendToOutgoingContext(forwardCtx, "x-jobconnect-internal", "job-service")
	if secret := strings.TrimSpace(os.Getenv("JOBCONNECT_INTERNAL_CALLER_SECRET")); secret != "" {
		forwardCtx = metadata.AppendToOutgoingContext(forwardCtx, "x-jobconnect-internal-secret", secret)
	}
	res, err := c.client.InternalGetJobOfferState(forwardCtx, &contractv1.GetJobOfferStateRequest{JobId: jobID})
	if err != nil {
		return application.ContractState{}, fmt.Errorf("get job offer state: %w", err)
	}
	return application.ContractState{
		JobID:             res.GetJobId(),
		HasPendingOffer:   res.GetHasPendingOffer(),
		PendingContractID: res.GetPendingContractId(),
		HasActiveContract: res.GetHasActiveContract(),
		ActiveContractID:  res.GetActiveContractId(),
	}, nil
}
