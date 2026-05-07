package events

import (
	"context"
	"encoding/json"
	"log"

	"jobconnect/job/internal/application"
	shared "jobconnect/events"

	"github.com/google/uuid"
)

type markFilledPayload struct {
	JobID    int64  `json:"job_id"`
	ClientID string `json:"client_id"`
}

func StartContractConsumer(ctx context.Context, brokers []string, topic string, markFilledUC *application.MarkJobFilled) *shared.Consumer {
	consumer := shared.NewConsumer(brokers, topic, "job-service")
	consumer.On("contract.job.mark_filled.requested", func(ctx context.Context, env shared.Envelope) error {
		var p markFilledPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil
		}
		clientID, err := uuid.Parse(p.ClientID)
		if err != nil {
			return nil
		}
		if _, err := markFilledUC.Execute(ctx, application.MarkJobFilledInput{JobID: p.JobID, ClientID: clientID}); err != nil {
			log.Printf("contract.job.mark_filled.requested failed: %v", err)
		}
		return nil
	})
	go func() { _ = consumer.Run(ctx) }()
	return consumer
}
