package applications

import "context"

type DeleteMessage struct {
	Chats ChatRepository
	Clock Clock
}

type DeleteMessageInput struct {
	MessageID string
	UserID    string
}

type DeleteMessageOutput struct {
	Success bool
}

func (dm *DeleteMessage) Execute(ctx context.Context, input DeleteMessageInput) (DeleteMessageOutput, error) {
	now := dm.Clock.Now()
	err := dm.Chats.DeleteMessage(ctx, input.MessageID, input.UserID, now)
	if err != nil {
		return DeleteMessageOutput{Success: false}, err
	}
	return DeleteMessageOutput{Success: true}, nil
}
