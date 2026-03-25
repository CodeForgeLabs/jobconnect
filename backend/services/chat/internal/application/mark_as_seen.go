package applications

import "context"

type MarkAsSeen struct {
	Chats ChatRepository
	Clock Clock
}

type MarkAsSeenInput struct {
	MessageID string
	UserID    string
}

type MarkAsSeenOutput struct {
	Success bool
}

func (ms *MarkAsSeen) Execute(ctx context.Context, input MarkAsSeenInput) (MarkAsSeenOutput, error) {
	now := ms.Clock.Now()
	err := ms.Chats.MarkAsSeen(ctx, input.MessageID, input.UserID, now)
	if err != nil {
		return MarkAsSeenOutput{Success: false}, err
	}
	return MarkAsSeenOutput{Success: true}, nil
}
