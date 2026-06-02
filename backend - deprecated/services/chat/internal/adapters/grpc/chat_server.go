package grpcadapter

import (
	"context"
	"errors"
	chatv1 "jobconnect/chat/gen/chat/v1"
	"jobconnect/chat/internal/adapters/ws"
	applications "jobconnect/chat/internal/application"
	"jobconnect/chat/internal/domain"
	"time"

	"google.golang.org/grpc/metadata"
)

type ChatServer struct {
	chatv1.UnimplementedChatServiceServer
	createMessage      *applications.CreateMessage
	getMessages        *applications.GetMessages
	markAsSeen         *applications.MarkAsSeen
	editMessage        *applications.EditMessage
	deleteMessage      *applications.DeleteMessage
	getConversation    *applications.GetConversation
	deleteConversation *applications.DeleteConversation
	hub                *ws.Hub
	TokenParser        TokenParser
}

func NewChatServer(createMessage *applications.CreateMessage, getMessages *applications.GetMessages, markAsSeen *applications.MarkAsSeen, editMessage *applications.EditMessage, deleteMessage *applications.DeleteMessage, getConversation *applications.GetConversation, deleteConversation *applications.DeleteConversation, hub *ws.Hub, tokenParser TokenParser) *ChatServer {
	return &ChatServer{createMessage: createMessage, getMessages: getMessages, markAsSeen: markAsSeen, editMessage: editMessage, deleteMessage: deleteMessage, getConversation: getConversation, deleteConversation: deleteConversation, hub: hub, TokenParser: tokenParser}
}

func (s *ChatServer) SendMessage(ctx context.Context, req *chatv1.SendMessageRequest) (*chatv1.SendMessageResponse, error) {
	input := applications.CreateMessageInput{
		SenderID:   req.GetSenderId(),
		ReceiverID: req.GetReceiverId(),
		Content:    getMessageContentFromProto(req.GetContent()),
	}

	output, err := s.createMessage.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	message := output.Message

	eventType := "NEW_MESSAGE"
	if output.IsNewConversation {
		eventType = "NEW_CONVERSATION"
	}

	// 3. Notify the Receiver
	s.hub.SendToUser(req.ReceiverId, map[string]interface{}{
		"event": eventType,
		"payload": map[string]interface{}{
			"message":       output.Message,
			"other_user_id": req.SenderId,
			"unseen_count":  1,
		},
	})
	return &chatv1.SendMessageResponse{
		Message: &chatv1.Message{
			Id:         message.ID,
			SenderId:   message.SenderID,
			ReceiverId: message.ReceiverID,
			Content:    mapDomainToProtoContent(message.Content),
			IsSeen:     message.IsSeen,
			SeenAt:     formatTime(message.SeenAt),
			IsEdited:   message.IsEdited,
			EditedAt:   formatTime(message.EditedAt),
			IsDeleted:  message.IsDeleted,
			DeletedAt:  formatTime(message.DeletedAt),
			CreatedAt:  message.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *ChatServer) GetMessages(ctx context.Context, req *chatv1.GetMessagesRequest) (*chatv1.GetMessagesResponse, error) {
	input := applications.GetMessagesInput{
		UserID1: req.GetUser1(),
		UserID2: req.GetUser2(),
	}

	output, err := s.getMessages.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	var messages []*chatv1.Message
	for _, msg := range output.Messages {
		messages = append(messages, &chatv1.Message{
			Id:         msg.ID,
			SenderId:   msg.SenderID,
			ReceiverId: msg.ReceiverID,
			Content:    mapDomainToProtoContent(msg.Content),
			IsSeen:     msg.IsSeen,
			SeenAt:     formatTime(msg.SeenAt),
			IsEdited:   msg.IsEdited,
			EditedAt:   formatTime(msg.EditedAt),
			IsDeleted:  msg.IsDeleted,
			DeletedAt:  formatTime(msg.DeletedAt),
			CreatedAt:  msg.CreatedAt.Format(time.RFC3339),
		})
	}

	return &chatv1.GetMessagesResponse{
		Messages: messages,
	}, nil
}

func (s *ChatServer) MarkAsSeen(ctx context.Context, req *chatv1.MarkAsSeenRequest) (*chatv1.MarkAsSeenResponse, error) {
	input := applications.MarkAsSeenInput{
		MessageID: req.GetMessageId(),
		UserID:    req.GetUserId(),
	}

	output, err := s.markAsSeen.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &chatv1.MarkAsSeenResponse{
		Success: output.Success,
	}, nil
}

func (s *ChatServer) EditMessage(ctx context.Context, req *chatv1.EditMessageRequest) (*chatv1.EditMessageResponse, error) {
	userId, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	input := applications.EditMessageInput{
		MessageID:  req.GetMessageId(),
		UserID:     userId, // later the userID should be extracted from the token
		NewContent: getMessageContentFromProto(req.GetNewContent()),
	}
	output, err := s.editMessage.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &chatv1.EditMessageResponse{
		Success: output.Success,
	}, nil
}

func (s *ChatServer) DeleteMessage(ctx context.Context, req *chatv1.DeleteMessageRequest) (*chatv1.DeleteMessageResponse, error) {
	userId, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	input := applications.DeleteMessageInput{
		MessageID: req.GetMessageId(),
		UserID:    userId, // later the userID should be extracted from the token
	}

	output, err := s.deleteMessage.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &chatv1.DeleteMessageResponse{
		Success: output.Success,
	}, nil
}
func (s *ChatServer) GetConversations(ctx context.Context, req *chatv1.GetConversationsRequest) (*chatv1.GetConversationsResponse, error) {
	input := applications.GetConversationInput{
		UserID: req.GetUserId(),
	}

	output, err := s.getConversation.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	var conversations []*chatv1.Conversation
	for _, conv := range output.Conversations {
		var lastMessage *chatv1.Message
		if (conv.LastMessage != domain.Message{}) {
			msg := conv.LastMessage
			lastMessage = &chatv1.Message{
				Id:         msg.ID,
				SenderId:   msg.SenderID,
				ReceiverId: msg.ReceiverID,
				Content:    mapDomainToProtoContent(msg.Content),
				IsSeen:     msg.IsSeen,
				SeenAt:     formatTime(msg.SeenAt),
				IsEdited:   msg.IsEdited,
				EditedAt:   formatTime(msg.EditedAt),
				IsDeleted:  msg.IsDeleted,
				DeletedAt:  formatTime(msg.DeletedAt),
				CreatedAt:  msg.CreatedAt.Format(time.RFC3339),
			}
		}
		conversations = append(conversations, &chatv1.Conversation{
			OtherUserId: conv.OtherUserID,
			LastMessage: lastMessage,
			UnseenCount: conv.UnseenCount,
		})
	}

	return &chatv1.GetConversationsResponse{
		Conversations: conversations,
	}, nil
}

func (s *ChatServer) DeleteConversation(ctx context.Context, req *chatv1.DeleteConversationRequest) (*chatv1.DeleteConversationResponse, error) {
	input := applications.DeleteConversationInput{
		UserID1: req.GetUser1(),
		UserID2: req.GetUser2(),
	}

	output, err := s.deleteConversation.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &chatv1.DeleteConversationResponse{
		Success: output.Success,
	}, nil
}

func getUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}

	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return "", errors.New("missing user_id")
	}

	userID := userIDs[0]
	if userID == "" {
		return "", errors.New("empty user_id")
	}

	return userID, nil
}
