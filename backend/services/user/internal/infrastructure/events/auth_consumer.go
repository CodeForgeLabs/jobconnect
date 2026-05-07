package events

import (
	"context"
	"encoding/json"
	"log"

	"jobconnect/user/internal/application"
	shared "jobconnect/events"

	"github.com/google/uuid"
)

type userRegisteredPayload struct {
	UserID      string `json:"user_id"`
	Role        string `json:"role"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

func StartAuthConsumer(ctx context.Context, brokers []string, topic string, createUC *application.CreateProfile) (*shared.Consumer, error) {
	consumer := shared.NewConsumer(brokers, topic, "user-service")
	consumer.On("auth.user.registered", func(ctx context.Context, env shared.Envelope) error {
		var p userRegisteredPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil
		}
		userID, err := uuid.Parse(p.UserID)
		if err != nil {
			return nil
		}
		_, err = createUC.Execute(ctx, application.CreateProfileInput{
			UserID:      userID,
			Role:        p.Role,
			FirstName:   p.FirstName,
			LastName:    p.LastName,
			DisplayName: p.DisplayName,
			AvatarURL:   "",
			ContactEmail: p.Email,
		})
		if err != nil {
			log.Printf("auth.user.registered handling failed user_id=%s err=%v", p.UserID, err)
		}
		return nil
	})
	go func() {
		if err := consumer.Run(ctx); err != nil {
			log.Printf("user auth consumer stopped: %v", err)
		}
	}()
	return consumer, nil
}
