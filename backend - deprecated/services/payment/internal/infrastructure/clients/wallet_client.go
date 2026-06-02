package clients

import (
	"context"
	"fmt"

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
	return fmt.Errorf("wallet RPC CreditWalletInternal is not available in current wallet.v1 API")
}

func (c *WalletClient) DebitWalletInternal(ctx context.Context, in application.DebitInput) error {
	return fmt.Errorf("wallet RPC DebitWalletInternal is not available in current wallet.v1 API")
}

func (c *WalletClient) PlaceHold(ctx context.Context, in application.PlaceHoldInput) (int64, error) {
	return 0, fmt.Errorf("wallet RPC PlaceHold is not available in current wallet.v1 API")
}

func (c *WalletClient) CaptureHold(ctx context.Context, in application.CaptureHoldInput) error {
	return fmt.Errorf("wallet RPC CaptureHold is not available in current wallet.v1 API")
}

func (c *WalletClient) GetBalance(ctx context.Context, walletID int64) (application.BalanceInfo, error) {
	return application.BalanceInfo{}, fmt.Errorf("wallet RPC GetBalance is not available in current wallet.v1 API")
}
