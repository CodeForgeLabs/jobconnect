package applications

import "context"

type DeleteConversation struct {
	Chats ChatRepository
}

type DeleteConversationInput struct {
	UserID1 string
	UserID2 string
}

type DeleteConversationOutput struct {
	Success bool
}

func (dc *DeleteConversation) Execute(ctx context.Context, input DeleteConversationInput) (DeleteConversationOutput, error) {
	err := dc.Chats.DeleteConversation(ctx, input.UserID1, input.UserID2)
	if err != nil {
		return DeleteConversationOutput{Success: false}, err
	}
	return DeleteConversationOutput{Success: true}, nil
}
