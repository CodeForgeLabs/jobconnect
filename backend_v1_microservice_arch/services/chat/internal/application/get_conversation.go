package applications

import (
	"context"
	"jobconnect/chat/internal/domain"
)

type GetConversation struct {
	Chats ChatRepository
}

type GetConversationInput struct {
	UserID string
}

type GetConversationOutput struct {
	Conversations []domain.Conversation
}

func (gc *GetConversation) Execute(ctx context.Context, input GetConversationInput) (GetConversationOutput, error) {
	conversations, err := gc.Chats.GetConversations(ctx, input.UserID)
	if err != nil {
		return GetConversationOutput{}, err
	}
	return GetConversationOutput{Conversations: conversations}, nil
}
