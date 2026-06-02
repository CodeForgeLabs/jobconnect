package applications

import (
	"context"
	"jobconnect/chat/internal/domain"
)

type CreateMessage struct {
	Chats ChatRepository
	Clock Clock
}
type CreateMessageInput struct {
	SenderID   string
	ReceiverID string
	Content    domain.MessageContent
}

type CreateMessageOutput struct {
	Message           domain.Message
	IsNewConversation bool
}

func (msg *CreateMessage) Execute(ctx context.Context, input CreateMessageInput) (CreateMessageOutput, error) {
	now := msg.Clock.Now()
	message := domain.Message{
		SenderID:   input.SenderID,
		ReceiverID: input.ReceiverID,
		Content:    input.Content,
		CreatedAt:  now,
	}
	newMessageOutput, err := msg.Chats.CreateMessage(ctx, message)
	if err != nil {
		return CreateMessageOutput{}, err
	}
	message.ID = newMessageOutput.MessageId
	return CreateMessageOutput{Message: message, IsNewConversation: newMessageOutput.IsNewConversation}, nil
}
