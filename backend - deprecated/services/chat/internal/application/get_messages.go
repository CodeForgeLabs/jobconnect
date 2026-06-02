package applications

import (
	"context"
	"jobconnect/chat/internal/domain"
)

type GetMessages struct {
	Chats ChatRepository
}

type GetMessagesInput struct {
	UserID1 string
	UserID2 string
}

type GetMessagesOutput struct {
	Messages []domain.Message
}

func (gm *GetMessages) Execute(ctx context.Context, input GetMessagesInput) (GetMessagesOutput, error) {
	messages, err := gm.Chats.GetMessages(ctx, input.UserID1, input.UserID2)
	if err != nil {
		return GetMessagesOutput{}, err
	}
	return GetMessagesOutput{Messages: messages}, nil
}
