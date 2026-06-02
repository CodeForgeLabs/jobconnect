package walletgrpc

import (
	"context"
	"time"

	walletv1 "jobconnect/dispute/gen/wallet/v1"
	"jobconnect/dispute/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type tokenIssuer interface {
	IssueAccessToken(userID uuid.UUID, role string, ttl time.Duration) (string, error)
}

const serviceActorID = "00000000-0000-0000-0000-00000000d15e"

type Client struct {
	grpc   walletv1.WalletServiceClient
	issuer tokenIssuer
}

func NewClient(grpc walletv1.WalletServiceClient, issuer tokenIssuer) *Client {
	return &Client{grpc: grpc, issuer: issuer}
}

func (c *Client) withAuth(ctx context.Context) (context.Context, error) {
	actorID, _ := uuid.Parse(serviceActorID)
	token, err := c.issuer.IssueAccessToken(actorID, "service", 2*time.Minute)
	if err != nil {
		return nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal", "dispute-service")
	return ctx, nil
}

func (c *Client) GetHoldByReference(ctx context.Context, referenceType, referenceID string) (application.Hold, error) {
	ctx, err := c.withAuth(ctx)
	if err != nil {
		return application.Hold{}, err
	}
	resp, err := c.grpc.GetHoldByReference(ctx, &walletv1.GetHoldByReferenceRequest{
		ReferenceType: referenceType,
		ReferenceId:   referenceID,
	})
	if err != nil {
		return application.Hold{}, err
	}
	hold := resp.GetHold()
	return application.Hold{
		ID:            hold.GetId(),
		WalletID:      hold.GetWalletId(),
		AmountMinor:   hold.GetAmountMinor(),
		CapturedMinor: hold.GetCapturedMinor(),
	}, nil
}

func (c *Client) ReleaseHold(ctx context.Context, holdID int64, idempotencyKey, note string) error {
	ctx, err := c.withAuth(ctx)
	if err != nil {
		return err
	}
	_, err = c.grpc.ReleaseHold(ctx, &walletv1.ReleaseHoldRequest{
		HoldId:         holdID,
		IdempotencyKey: idempotencyKey,
		Note:           note,
	})
	return err
}

func (c *Client) CaptureHold(ctx context.Context, holdID, amountMinor int64, idempotencyKey, referenceType, referenceID, note string) error {
	ctx, err := c.withAuth(ctx)
	if err != nil {
		return err
	}
	_, err = c.grpc.CaptureHold(ctx, &walletv1.CaptureHoldRequest{
		HoldId:             holdID,
		CaptureAmountMinor: amountMinor,
		IdempotencyKey:     idempotencyKey,
		ReferenceType:      referenceType,
		ReferenceId:        referenceID,
		Note:               note,
	})
	return err
}
