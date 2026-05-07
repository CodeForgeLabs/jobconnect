package events

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"jobconnect/services/connects/internal/application"
	shared "jobconnect/events"
)

type userRegisteredPayload struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func StartAuthConsumer(ctx context.Context, brokers []string, topic string, uc *application.UseCases) *shared.Consumer {
	consumer := shared.NewConsumer(brokers, topic, "connects-service")
	consumer.On("auth.user.registered", func(ctx context.Context, env shared.Envelope) error {
		var p userRegisteredPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil
		}
		if strings.ToLower(strings.TrimSpace(p.Role)) != "freelancer" {
			return nil
		}
		if _, err := uc.GrantInitialConnects(ctx, p.UserID); err != nil {
			log.Printf("grant initial connects failed user_id=%s err=%v", p.UserID, err)
		}
		return nil
	})
	go func() {
		if err := consumer.Run(ctx); err != nil {
			log.Printf("connects auth consumer stopped: %v", err)
		}
	}()
	return consumer
}
