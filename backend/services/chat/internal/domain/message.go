package domain

import (
	"time"
)

// MessageType helps identify which part of the 'oneof' is active
type MessageType string

const (
	TypeText  MessageType = "text"
	TypeImage MessageType = "image"
	TypeVideo MessageType = "video"
)

// MessageContent represents the data within the message
type MessageContent struct {
	Type     MessageType
	Text     string
	ImageUrl string
	VideoUrl string
	Caption  string // Used for both image and video
}

// Message is the core entity for your business logic
type Message struct {
	ID         string
	SenderID   string
	ReceiverID string
	Content    MessageContent

	IsSeen bool
	SeenAt *time.Time // Using pointer for nullability

	IsEdited bool
	EditedAt *time.Time

	IsDeleted bool
	DeletedAt *time.Time

	CreatedAt time.Time
}

// Conversation represents a single chat entry in a user's inbox
type Conversation struct {
	OtherUserID string
	LastMessage Message
	UnseenCount int32
}

type NewMessageOutput struct {
	MessageId         string
	IsNewConversation bool
}
