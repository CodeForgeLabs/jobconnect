package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	chatv1 "jobconnect/chat/gen/chat/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

type ChatHandler struct {
	client chatv1.ChatServiceClient
}

type SendMessageRequest struct {
	ReceiverID string      `json:"receiver_id"`
	Content    interface{} `json:"content"`
}

type EditMessageRequest struct {
	NewContent interface{} `json:"new_content"`
}

type ChatMessageDTO struct {
	ID         string      `json:"id"`
	SenderID   string      `json:"sender_id"`
	ReceiverID string      `json:"receiver_id"`
	Content    interface{} `json:"content"`
	IsSeen     bool        `json:"is_seen"`
	SeenAt     interface{} `json:"seen_at"`
	IsEdited   bool        `json:"is_edited"`
	EditedAt   interface{} `json:"edited_at"`
	IsDeleted  bool        `json:"is_deleted"`
	DeletedAt  interface{} `json:"deleted_at"`
	CreatedAt  interface{} `json:"created_at"`
}

type ConversationDTO struct {
	OtherUserID string         `json:"other_user_id"`
	LastMessage ChatMessageDTO `json:"last_message"`
	UnseenCount uint32         `json:"unseen_count"`
}

type SendMessageResponse struct {
	Message ChatMessageDTO `json:"message"`
}

type GetMessagesResponse struct {
	Messages []ChatMessageDTO `json:"messages"`
}

type GetConversationsResponse struct {
	Conversations []ConversationDTO `json:"conversations"`
}

type ChatStatusResponse struct {
	Status string `json:"status"`
}

type ChatErrorResponse struct {
	Error string `json:"error"`
}

func NewChatHandler(client chatv1.ChatServiceClient) *ChatHandler {
	return &ChatHandler{client: client}
}

// SendMessage godoc
// @Summary Send a chat message
// @Description Sends a new message from the authenticated user to another user.
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SendMessageRequest true "Message payload"
// @Success 200 {object} SendMessageResponse
// @Failure 400 {object} ChatErrorResponse
// @Failure 401 {object} ChatErrorResponse
// @Failure 500 {object} ChatErrorResponse
// @Router /api/v1/chat/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	ctx := attachUserID(c.Request.Context(), userID)

	var body struct {
		ReceiverID string          `json:"receiver_id"`
		Content    json.RawMessage `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.ReceiverID == "" || body.Content == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "receiver_id and content are required"})
		return
	}
	var protoContent chatv1.MessageContent

	if err := protojson.Unmarshal(body.Content, &protoContent); err != nil {
		return
	}
	resp, err := h.client.SendMessage(ctx, &chatv1.SendMessageRequest{
		SenderId:   userID,
		ReceiverId: body.ReceiverID,
		Content:    &protoContent,
	})

	if err != nil {
		writeGRPCError(c, err)
		return
	}

	msg := resp.GetMessage()

	c.JSON(http.StatusOK, gin.H{
		"message": mapProtoToHTTPMessage(msg),
	})
}

// GetMessages godoc
// @Summary Get messages with a user
// @Description Returns all messages in the conversation between the authenticated user and the specified user.
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param userId path string true "Other user ID"
// @Success 200 {object} GetMessagesResponse
// @Failure 400 {object} ChatErrorResponse
// @Failure 401 {object} ChatErrorResponse
// @Failure 500 {object} ChatErrorResponse
// @Router /api/v1/chat/{userId}/messages [get]
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	ctx := attachUserID(c.Request.Context(), userID)

	otherUserID := strings.TrimSpace(c.Param("userId"))
	if otherUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
		return
	}

	resp, err := h.client.GetMessages(ctx, &chatv1.GetMessagesRequest{
		User1: userID,
		User2: otherUserID,
	})

	if err != nil {
		writeGRPCError(c, err)
		return
	}
	msgs := resp.GetMessages()

	var messages []map[string]interface{}
	for _, msg := range msgs {
		messages = append(messages, mapProtoToHTTPMessage(msg))
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// GetMyConversations godoc
// @Summary List my conversations
// @Description Returns conversation summaries for the authenticated user.
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Success 200 {object} GetConversationsResponse
// @Failure 401 {object} ChatErrorResponse
// @Failure 500 {object} ChatErrorResponse
// @Router /api/v1/chat/conversations [get]
func (h *ChatHandler) GetMyConversations(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	ctx := attachUserID(c.Request.Context(), userID)

	resp, err := h.client.GetConversations(ctx, &chatv1.GetConversationsRequest{
		UserId: userID,
	})

	if err != nil {
		writeGRPCError(c, err)
		return
	}
	conversations := resp.GetConversations()

	var convs []map[string]interface{}
	for _, conv := range conversations {
		convs = append(convs, map[string]interface{}{
			"other_user_id": conv.GetOtherUserId(),
			"last_message":  mapProtoToHTTPMessage(conv.GetLastMessage()),
			"unseen_count":  conv.GetUnseenCount(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"conversations": convs})
}

// MarkAsSeen godoc
// @Summary Mark a message as seen
// @Description Marks a specific message as seen by the authenticated user.
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param messageId path string true "Message ID"
// @Success 200 {object} ChatStatusResponse
// @Failure 401 {object} ChatErrorResponse
// @Failure 500 {object} ChatErrorResponse
// @Router /api/v1/chat/messages/{messageId}/seen [post]
func (h *ChatHandler) MarkAsSeen(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	ctx := attachUserID(c.Request.Context(), userID)

	messageID := c.Param("messageId")

	_, err := h.client.MarkAsSeen(ctx, &chatv1.MarkAsSeenRequest{
		UserId:    userID,
		MessageId: messageID,
	})

	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "seen"})
}

// EditMessage godoc
// @Summary Edit a message
// @Description Updates content of an existing message.
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param messageId path string true "Message ID"
// @Param request body EditMessageRequest true "Edit message payload"
// @Success 200 {object} ChatStatusResponse
// @Failure 400 {object} ChatErrorResponse
// @Failure 401 {object} ChatErrorResponse
// @Failure 500 {object} ChatErrorResponse
// @Router /api/v1/chat/messages/{messageId} [put]
func (h *ChatHandler) EditMessage(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	ctx := attachUserID(c.Request.Context(), userID)
	messageID := c.Param("messageId")

	var body struct {
		NewContent json.RawMessage `json:"new_content"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(body.NewContent) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new_content required"})
		return
	}

	var protoContent chatv1.MessageContent

	if err := protojson.Unmarshal(body.NewContent, &protoContent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.client.EditMessage(ctx, &chatv1.EditMessageRequest{
		MessageId:  messageID,
		NewContent: &protoContent,
	})

	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "edited"})
}

// DeleteMessage godoc
// @Summary Delete a message
// @Description Deletes a specific message.
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param messageId path string true "Message ID"
// @Success 200 {object} ChatStatusResponse
// @Failure 401 {object} ChatErrorResponse
// @Failure 500 {object} ChatErrorResponse
// @Router /api/v1/chat/messages/{messageId} [delete]
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	ctx := attachUserID(c.Request.Context(), userID)

	messageID := c.Param("messageId")

	_, err := h.client.DeleteMessage(ctx, &chatv1.DeleteMessageRequest{
		MessageId: messageID,
	})

	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// DeleteConversation godoc
// @Summary Delete conversation
// @Description Deletes all messages in the conversation between authenticated user and specified user.
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param userId path string true "Other user ID"
// @Success 200 {object} ChatStatusResponse
// @Failure 400 {object} ChatErrorResponse
// @Failure 401 {object} ChatErrorResponse
// @Failure 500 {object} ChatErrorResponse
// @Router /api/v1/chat/conversations/{userId} [delete]
func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	ctx := attachUserID(c.Request.Context(), userID)

	otherUserID := strings.TrimSpace(c.Param("userId"))
	if otherUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
		return
	}

	_, err := h.client.DeleteConversation(ctx, &chatv1.DeleteConversationRequest{
		User1: userID,
		User2: otherUserID,
	})

	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "conversation deleted"})
}

func mapProtoToHTTPMessage(m *chatv1.Message) map[string]interface{} {
	return map[string]interface{}{
		"id":          m.Id,
		"sender_id":   m.SenderId,
		"receiver_id": m.ReceiverId,
		"content":     m.Content,
		"is_seen":     m.IsSeen,
		"seen_at":     m.SeenAt,
		"is_edited":   m.IsEdited,
		"edited_at":   m.EditedAt,
		"is_deleted":  m.IsDeleted,
		"deleted_at":  m.DeletedAt,
		"created_at":  m.CreatedAt,
	}
}
