package clients

import (
	"context"
	"fmt"
	"os"
	"strings"

	shared "jobconnect/events"
	contractv1 "jobconnect/payment/gen/contract/v1"
	"jobconnect/payment/internal/application"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type ContractClient struct {
	client contractv1.ContractServiceClient
	conn   *grpc.ClientConn
	events *shared.Publisher
}

func NewContractClient(addr string) (*ContractClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to contract service: %w", err)
	}
	return &ContractClient{
		client: contractv1.NewContractServiceClient(conn),
		conn:   conn,
		events: shared.NewPublisher(shared.ParseBrokers(os.Getenv("KAFKA_BROKERS")), getEnv("KAFKA_TOPIC_PAYMENT", "payment.events")),
	}, nil
}

func (c *ContractClient) Close() error {
	if c.events != nil {
		_ = c.events.Close()
	}
	return c.conn.Close()
}

func (c *ContractClient) MarkMilestoneFunded(ctx context.Context, contractID int64, milestoneID int64) error {
	if c.events != nil {
		if env, err := shared.NewEnvelope("payment.contract.mark_milestone_funded.requested", fmt.Sprintf("%d:%d", contractID, milestoneID), "payment-service", 1, map[string]any{"contract_id": contractID, "milestone_id": milestoneID}, fmt.Sprintf("milestone-funded:%d:%d", contractID, milestoneID), fmt.Sprintf("%d", contractID)); err == nil {
			_ = c.events.Publish(ctx, env)
		}
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal", "payment-service")
	if secret := strings.TrimSpace(os.Getenv("JOBCONNECT_INTERNAL_CALLER_SECRET")); secret != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal-secret", secret)
	}
	_, err := c.client.InternalMarkMilestoneFunded(ctx, &contractv1.InternalMarkMilestoneFundedRequest{
		ContractId:  contractID,
		MilestoneId: milestoneID,
	})
	if err != nil {
		return fmt.Errorf("mark milestone funded: %w", err)
	}
	return nil
}

func (c *ContractClient) MarkContractBonusPaid(ctx context.Context, bonusID int64, paymentReferenceID string) error {
	if c.events != nil {
		if env, err := shared.NewEnvelope("payment.contract.mark_bonus_paid.requested", fmt.Sprintf("%d", bonusID), "payment-service", 1, map[string]any{"bonus_id": bonusID, "payment_reference_id": paymentReferenceID}, fmt.Sprintf("bonus-paid:%d", bonusID), fmt.Sprintf("%d", bonusID)); err == nil {
			_ = c.events.Publish(ctx, env)
		}
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal", "payment-service")
	if secret := strings.TrimSpace(os.Getenv("JOBCONNECT_INTERNAL_CALLER_SECRET")); secret != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal-secret", secret)
	}
	paymentReferenceID = strings.TrimSpace(paymentReferenceID)
	if paymentReferenceID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-payment-reference-id", paymentReferenceID)
	}
	_, err := c.client.InternalMarkContractBonusPaid(ctx, &contractv1.InternalMarkContractBonusPaidRequest{
		BonusId: bonusID,
	})
	if err != nil {
		return fmt.Errorf("mark contract bonus paid: %w", err)
	}
	return nil
}

func (c *ContractClient) CloseHourlyWeek(ctx context.Context, contractID int64, weekStartUnixSeconds int64) error {
	if c.events != nil {
		if env, err := shared.NewEnvelope("payment.contract.close_hourly_week.requested", fmt.Sprintf("%d", contractID), "payment-service", 1, map[string]any{"contract_id": contractID, "week_start_unix_seconds": weekStartUnixSeconds}, fmt.Sprintf("close-hourly-week:%d:%d", contractID, weekStartUnixSeconds), fmt.Sprintf("%d", contractID)); err == nil {
			_ = c.events.Publish(ctx, env)
		}
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal", "payment-service")
	if secret := strings.TrimSpace(os.Getenv("JOBCONNECT_INTERNAL_CALLER_SECRET")); secret != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal-secret", secret)
	}
	_, err := c.client.InternalCloseHourlyWeek(ctx, &contractv1.InternalCloseHourlyWeekRequest{
		ContractId:           contractID,
		WeekStartUnixSeconds: weekStartUnixSeconds,
	})
	if err != nil {
		return fmt.Errorf("close hourly week: %w", err)
	}
	return nil
}

func (c *ContractClient) SettleHourlyInvoice(ctx context.Context, invoiceID int64) error {
	if c.events != nil {
		if env, err := shared.NewEnvelope("payment.contract.settle_hourly_invoice.requested", fmt.Sprintf("%d", invoiceID), "payment-service", 1, map[string]any{"invoice_id": invoiceID}, fmt.Sprintf("settle-hourly-invoice:%d", invoiceID), fmt.Sprintf("%d", invoiceID)); err == nil {
			_ = c.events.Publish(ctx, env)
		}
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal", "payment-service")
	if secret := strings.TrimSpace(os.Getenv("JOBCONNECT_INTERNAL_CALLER_SECRET")); secret != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal-secret", secret)
	}
	_, err := c.client.InternalSettleHourlyInvoice(ctx, &contractv1.InternalSettleHourlyInvoiceRequest{
		InvoiceId: invoiceID,
	})
	if err != nil {
		return fmt.Errorf("settle hourly invoice: %w", err)
	}
	return nil
}

var _ application.ContractClient = (*ContractClient)(nil)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
