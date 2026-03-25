package db

import (
	"context"
	"jobconnect/chat/internal/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepo struct {
	pool *pgxpool.Pool
}

func NewChatRepo(pool *pgxpool.Pool) *ChatRepo {
	return &ChatRepo{pool: pool}
}

func (r *ChatRepo) CreateMessage(ctx context.Context, message domain.Message) (domain.NewMessageOutput, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.NewMessageOutput{}, err
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM messages 
            WHERE (sender_id = $1 AND receiver_id = $2) 
               OR (sender_id = $2 AND receiver_id = $1)
            LIMIT 1
        )
    `, message.SenderID, message.ReceiverID).Scan(&exists)

	if err != nil {
		return domain.NewMessageOutput{}, err
	}

	isNewConversation := !exists

	var id string
	err = tx.QueryRow(ctx, `
        INSERT INTO messages (
            sender_id, receiver_id, type, text, image_url, video_url, caption,
            is_seen, is_edited, is_deleted, created_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id
    `,
		message.SenderID, message.ReceiverID, message.Content.Type,
		message.Content.Text, message.Content.ImageUrl, message.Content.VideoUrl, message.Content.Caption,
		false, false, false, message.CreatedAt,
	).Scan(&id)

	if err != nil {
		return domain.NewMessageOutput{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.NewMessageOutput{}, err
	}

	return domain.NewMessageOutput{
		MessageId:         id,
		IsNewConversation: isNewConversation,
	}, nil
}

func (r *ChatRepo) GetMessages(ctx context.Context, userID1, userID2 string) ([]domain.Message, error) {
	rows, err := r.pool.Query(ctx, `
		select id, sender_id, receiver_id, type, text, image_url, video_url, caption,
		       is_seen, seen_at, is_edited, edited_at, is_deleted, deleted_at, created_at
		from messages
		where (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
	`, userID1, userID2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		var contentType string
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &contentType,
			&msg.Content.Text, &msg.Content.ImageUrl, &msg.Content.VideoUrl, &msg.Content.Caption,
			&msg.IsSeen, &msg.SeenAt, &msg.IsEdited, &msg.EditedAt, &msg.IsDeleted, &msg.DeletedAt, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		msg.Content.Type = domain.MessageType(contentType)
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *ChatRepo) MarkAsSeen(ctx context.Context, messageID string, userId string, seenAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
        UPDATE messages
        SET is_seen = true, seen_at = $3
        WHERE id = $1 
          AND receiver_id = $2 
          AND is_seen = false
    `, messageID, userId, seenAt)

	return err
}

func (r *ChatRepo) EditMessage(ctx context.Context, messageID string, userId string, newContent domain.MessageContent, editedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
        UPDATE messages
        SET 
            -- If the new value is empty/null, keep the existing column value
            text = CASE WHEN $3 = '' THEN text ELSE $3 END,
            image_url = CASE WHEN $4 = '' THEN image_url ELSE $4 END,
            video_url = CASE WHEN $5 = '' THEN video_url ELSE $5 END,
            caption = CASE WHEN $6 = '' THEN caption ELSE $6 END,
            
            -- Always update these when an edit happens
            is_edited = true, 
            edited_at = $7
        WHERE id = $1 AND sender_id = $2
    `,
		messageID,
		userId,
		newContent.Text,
		newContent.ImageUrl,
		newContent.VideoUrl,
		newContent.Caption,
		editedAt)

	return err
}

func (r *ChatRepo) DeleteMessage(ctx context.Context, messageID string, userId string, deletedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE messages
		SET 
		    is_deleted = true, 
		    deleted_at = $3
		WHERE id = $1 AND sender_id = $2
	`, messageID, userId, deletedAt)

	return err
}

func (r *ChatRepo) GetConversations(ctx context.Context, userID string) ([]domain.Conversation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 
		    CASE 
		        WHEN sender_id = $1 THEN receiver_id 
		        ELSE sender_id 
		    END AS other_user_id,
		    id, sender_id, receiver_id, type, text, image_url, video_url, caption,
		    is_seen, seen_at, is_edited, edited_at, is_deleted, deleted_at, created_at
		FROM messages
		WHERE sender_id = $1 OR receiver_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversationMap := make(map[string]domain.Conversation)
	for rows.Next() {
		var otherUserID string
		var msg domain.Message
		var contentType string
		err := rows.Scan(&otherUserID, &msg.ID, &msg.SenderID, &msg.ReceiverID, &contentType,
			&msg.Content.Text, &msg.Content.ImageUrl, &msg.Content.VideoUrl, &msg.Content.Caption,
			&msg.IsSeen, &msg.SeenAt, &msg.IsEdited, &msg.EditedAt, &msg.IsDeleted, &msg.DeletedAt, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		msg.Content.Type = domain.MessageType(contentType)

		if conv, exists := conversationMap[otherUserID]; exists {
			if msg.CreatedAt.After(conv.LastMessage.CreatedAt) {
				conv.LastMessage = msg
				conversationMap[otherUserID] = conv
			}
			if !msg.IsSeen && msg.ReceiverID == userID {
				conv.UnseenCount++
				conversationMap[otherUserID] = conv
			}
		} else {
			unseenCount := int32(0)
			if !msg.IsSeen && msg.ReceiverID == userID {
				unseenCount = 1
			}
			conversationMap[otherUserID] = domain.Conversation{
				OtherUserID: otherUserID,
				LastMessage: msg,
				UnseenCount: unseenCount,
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var conversations []domain.Conversation
	for _, conv := range conversationMap {
		conversations = append(conversations, conv)
	}
	return conversations, nil
}

func (r *ChatRepo) DeleteConversation(ctx context.Context, userID1, userID2 string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM messages
		WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
	`, userID1, userID2)
	return err
}
