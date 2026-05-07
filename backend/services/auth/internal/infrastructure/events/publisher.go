package events

import (
	"context"

	"jobconnect/auth/internal/application"
	shared "jobconnect/events"
)

type UserRegistrationPublisher struct {
	publisher *shared.Publisher
}

type userRegisteredPayload struct {
	UserID      string `json:"user_id"`
	Role        string `json:"role"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

func NewUserRegistrationPublisher(p *shared.Publisher) *UserRegistrationPublisher {
	return &UserRegistrationPublisher{publisher: p}
}

func (p *UserRegistrationPublisher) PublishUserRegistered(ctx context.Context, in application.CreateProfileInput) error {
	env, err := shared.NewEnvelope(
		"auth.user.registered",
		in.UserID.String(),
		"auth-service",
		1,
		userRegisteredPayload{
			UserID:      in.UserID.String(),
			Role:        in.Role,
			FirstName:   in.FirstName,
			LastName:    in.LastName,
			DisplayName: in.DisplayName,
			Email:       in.Email,
		},
		"register:"+in.UserID.String(),
		in.UserID.String(),
	)
	if err != nil {
		return err
	}
	return p.publisher.Publish(ctx, env)
}
