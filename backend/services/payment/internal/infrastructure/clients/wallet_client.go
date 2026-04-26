package clients

import (
	"context"
	"fmt"
	"time"

	"jobconnect/payment/internal/application"
	walletv1 "jobconnect/wallet/gen/wallet/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WalletClient struct {
	client walletv1.WalletServiceClient
	conn   *grpc.ClientConn
}

func NewWalletClient(addr string) (*WalletClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to wallet service: %w", err)
	}

	return &WalletClient{
		client: walletv1.NewWalletServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *WalletClient) Close() error {
	return c.conn.Close()
}

// Ensure WalletClient implements application.WalletClient
var _ application.WalletClient = (*WalletClient)(nil)

func (c *WalletClient) CreditWalletInternal(ctx context.Context, in application.CreditInput) error {
	req := &walletv1.CreditWalletInternalRequest{
		WalletId:       in.WalletID,
		AmountMinor:    in.AmountMinor,
		IdempotencyKey: in.IdempotencyKey,
		ReferenceType:  in.ReferenceType,
		ReferenceId:    in.ReferenceID,
		Note:           in.Note,
	}

	_, err := c.client.CreditWalletInternal(ctx, req)
	if err != nil {
		return fmt.Errorf("credit failed: %w", err)
	}

	return nil
}

func (c *WalletClient) DebitWalletInternal(ctx context.Context, in application.DebitInput) error {
	req := &walletv1.DebitWalletInternalRequest{
		WalletId:       in.WalletID,
		AmountMinor:    in.AmountMinor,
		IdempotencyKey: in.IdempotencyKey,
		ReferenceType:  in.ReferenceType,
		ReferenceId:    in.ReferenceID,
		Note:           in.Note,
	}

	_, err := c.client.DebitWalletInternal(ctx, req)
	if err != nil {
		return fmt.Errorf("debit failed: %w", err)
	}

	return nil
}

func (c *WalletClient) PlaceHold(ctx context.Context, in application.PlaceHoldInput) (int64, error) {
	req := &walletv1.PlaceHoldRequest{
		WalletId:             in.WalletID,
		AmountMinor:          in.AmountMinor,
		IdempotencyKey:       in.IdempotencyKey,
		ReferenceType:        in.ReferenceType,
		ReferenceId:          in.ReferenceID,
		Note:                 in.Note,
		ExpiresAtUnixSeconds: time.Now().Add(30 * 24 * time.Hour).Unix(), // Arbitrary default
	}

	res, err := c.client.PlaceHold(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("place hold failed: %w", err)
	}

	return res.Hold.Id, nil
}

func (c *WalletClient) CaptureHold(ctx context.Context, in application.CaptureHoldInput) error {
	req := &walletv1.CaptureHoldRequest{
		HoldId:             in.HoldID,
		CaptureAmountMinor: in.CaptureAmountMinor,
		IdempotencyKey:     in.IdempotencyKey,
		ReferenceType:      in.ReferenceType,
		ReferenceId:        in.ReferenceID,
		Note:               in.Note,
	}

	_, err := c.client.CaptureHold(ctx, req)
	if err != nil {
		return fmt.Errorf("capture hold failed: %w", err)
	}

	return nil
}

func (c *WalletClient) GetBalance(ctx context.Context, walletID int64) (application.BalanceInfo, error) {
	req := &walletv1.GetBalanceRequest{
		WalletId: walletID,
	}

	res, err := c.client.GetBalance(ctx, req)
	if err != nil {
		return application.BalanceInfo{}, fmt.Errorf("get balance failed: %w", err)
	}

	return application.BalanceInfo{
		AvailableMinor: res.Balance.AvailableMinor,
		HeldMinor:      res.Balance.HeldMinor,
		TotalMinor:     res.Balance.TotalMinor,
	}, nil
}
