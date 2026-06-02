package applications

import (
	"context"
	"jobconnect/chat/internal/domain"
	"time"
)

type ChatRepository interface {
	CreateMessage(ctx context.Context, message domain.Message) (domain.NewMessageOutput, error)
	GetMessages(ctx context.Context, userID1, userID2 string) ([]domain.Message, error)
	MarkAsSeen(ctx context.Context, messageID string, userId string, seenAt time.Time) error
	EditMessage(ctx context.Context, messageID string, userId string, newContent domain.MessageContent, editedAt time.Time) error
	DeleteMessage(ctx context.Context, messageID string, userId string, deletedAt time.Time) error
	GetConversations(ctx context.Context, userID string) ([]domain.Conversation, error)
	DeleteConversation(ctx context.Context, userID1, userID2 string) error
}

type Clock interface {
	Now() time.Time
}
