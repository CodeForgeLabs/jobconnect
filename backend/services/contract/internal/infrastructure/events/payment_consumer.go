package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"jobconnect/contract/internal/application"
	"jobconnect/contract/internal/domain"
	shared "jobconnect/events"
)

type contractPaymentPayload struct {
	ContractID           int64  `json:"contract_id"`
	MilestoneID          int64  `json:"milestone_id"`
	BonusID              int64  `json:"bonus_id"`
	PaymentReferenceID   string `json:"payment_reference_id"`
	WeekStartUnixSeconds int64  `json:"week_start_unix_seconds"`
	InvoiceID            int64  `json:"invoice_id"`
}

func StartPaymentConsumer(ctx context.Context, brokers []string, topic string, updateMilestoneUC *application.UpdateMilestoneStatus, markBonusUC *application.InternalMarkContractBonusPaid, closeWeekUC *application.InternalCloseHourlyWeek, settleUC *application.InternalSettleHourlyInvoice) *shared.Consumer {
	consumer := shared.NewConsumer(brokers, topic, "contract-service")
	consumer.On("payment.contract.mark_bonus_paid.requested", func(ctx context.Context, env shared.Envelope) error {
		var p contractPaymentPayload
		_ = json.Unmarshal(env.Payload, &p)
		_, err := markBonusUC.Execute(ctx, application.InternalMarkContractBonusPaidInput{BonusID: p.BonusID, PaymentReferenceID: p.PaymentReferenceID})
		if err != nil {
			log.Printf("payment.contract.mark_bonus_paid.requested failed: %v", err)
		}
		return nil
	})
	consumer.On("payment.contract.mark_milestone_funded.requested", func(ctx context.Context, env shared.Envelope) error {
		var p contractPaymentPayload
		_ = json.Unmarshal(env.Payload, &p)
		serviceID := "00000000-0000-0000-0000-00000000c0de"
		actorID, _ := uuid.Parse(serviceID)
		_, err := updateMilestoneUC.Execute(ctx, application.UpdateMilestoneStatusInput{
			ContractID:  p.ContractID,
			MilestoneID: p.MilestoneID,
			ActorID:     actorID,
			ActorRole:   "internal",
			Status:      domain.MilestoneStatusFunded,
		})
		if err != nil {
			log.Printf("payment.contract.mark_milestone_funded.requested failed: %v", err)
		}
		return nil
	})
	consumer.On("payment.contract.close_hourly_week.requested", func(ctx context.Context, env shared.Envelope) error {
		var p contractPaymentPayload
		_ = json.Unmarshal(env.Payload, &p)
		weekStart := time.Unix(p.WeekStartUnixSeconds, 0).UTC()
		if p.WeekStartUnixSeconds <= 0 {
			weekStart = time.Now().UTC()
		}
		_, err := closeWeekUC.Execute(ctx, application.InternalCloseHourlyWeekInput{ContractID: p.ContractID, WeekStart: weekStart})
		if err != nil {
			log.Printf("payment.contract.close_hourly_week.requested failed: %v", err)
		}
		return nil
	})
	consumer.On("payment.contract.settle_hourly_invoice.requested", func(ctx context.Context, env shared.Envelope) error {
		var p contractPaymentPayload
		_ = json.Unmarshal(env.Payload, &p)
		_, err := settleUC.Execute(ctx, application.InternalSettleHourlyInvoiceInput{InvoiceID: p.InvoiceID})
		if err != nil {
			log.Printf("payment.contract.settle_hourly_invoice.requested failed: %v", err)
		}
		return nil
	})
	go func() { _ = consumer.Run(ctx) }()
	return consumer
}
