package applications

import (
	"context"
	"jobconnect/chat/internal/domain"
)

type EditMessage struct {
	Chats ChatRepository
	Clock Clock
}

type EditMessageInput struct {
	MessageID  string
	UserID     string
	NewContent domain.MessageContent
}

type EditMessageOutput struct {
	Success bool
}

func (em *EditMessage) Execute(ctx context.Context, input EditMessageInput) (EditMessageOutput, error) {
	now := em.Clock.Now()
	err := em.Chats.EditMessage(ctx, input.MessageID, input.UserID, input.NewContent, now)
	if err != nil {
		return EditMessageOutput{Success: false}, err
	}
	return EditMessageOutput{Success: true}, nil
}
