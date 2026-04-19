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

func NewChatHandler(client chatv1.ChatServiceClient) *ChatHandler {
	return &ChatHandler{client: client}
}

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
