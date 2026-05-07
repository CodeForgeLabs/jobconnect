package events

import (
	"context"
	"encoding/json"
	"log"

	shared "jobconnect/events"
	"jobconnect/recommendation/internal/application"
)

type reviewPayload struct {
	RevieweeID string `json:"reviewee_id"`
}

func StartReviewConsumer(ctx context.Context, brokers []string, topic string, svc *application.RecommendationService) *shared.Consumer {
	consumer := shared.NewConsumer(brokers, topic, "recommendation-service")
	handler := func(ctx context.Context, env shared.Envelope) error {
		var p reviewPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil
		}
		if _, err := svc.InvalidateRecommendationCache(ctx, []string{p.RevieweeID}, nil, false); err != nil {
			log.Printf("recommendation invalidation from review event failed: %v", err)
		}
		return nil
	}
	consumer.On("review.created", handler)
	consumer.On("review.updated", handler)
	consumer.On("review.deleted", handler)
	go func() {
		if err := consumer.Run(ctx); err != nil {
			log.Printf("recommendation review consumer stopped: %v", err)
		}
	}()
	return consumer
}
