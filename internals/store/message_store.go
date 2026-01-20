package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeFile  MessageType = "file"
)

type Message struct {
	ID               uuid.UUID   `json:"id" db:"id"`
	ConversationID   uuid.UUID   `json:"conversation_id" db:"conversation_id"`
	SenderID         uuid.UUID   `json:"sender_id" db:"sender_id"`
	Content          string      `json:"content" db:"content"`
	MessageType      MessageType `json:"message_type" db:"message_type"`
	MediaURL         *string     `json:"media_url,omitempty" db:"media_url"`
	MediaSize        *int64      `json:"media_size,omitempty" db:"media_size"`
	MediaMimeType    *string     `json:"media_mime_type,omitempty" db:"media_mime_type"`
	ReplyToMessageID *uuid.UUID  `json:"reply_to_message_id,omitempty" db:"reply_to_message_id"`
	CreatedAt        time.Time   `json:"created_at" db:"created_at"`
	EditedAt         *time.Time  `json:"edited_at,omitempty" db:"edited_at"`
	DeletedAt        *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
}

type MessageStatusType string

const (
	MessageStatusSent      MessageStatusType = "sent"
	MessageStatusDelivered MessageStatusType = "delivered"
	MessageStatusRead      MessageStatusType = "read"
)

type MessageStatus struct {
	ID        uuid.UUID         `json:"id" db:"id"`
	MessageID uuid.UUID         `json:"message_id" db:"message_id"`
	UserID    uuid.UUID         `json:"user_id" db:"user_id"`
	Status    MessageStatusType `json:"status" db:"status"`
	Timestamp time.Time         `json:"timestamp" db:"timestamp"`
}

type ConversationWithDetails struct {
	Conversation
	ParticipantCount int      `json:"participant_count"`
	LastMessage      *Message `json:"last_message,omitempty"`
	UnreadCount      int      `json:"unread_count"`
}

type PostgresMessageStore struct {
	DB *sql.DB
}

func NewPostgresMessageStore(db *sql.DB) *PostgresMessageStore {
	return &PostgresMessageStore{
		DB: db,
	}
}

type MessageStore interface {
	CreateMessage(ctx context.Context, msg *Message) (*Message, error)
	GetMessagesByConversationID(ctx context.Context, conversationID uuid.UUID, limit int, before *time.Time) ([]Message, error)
	CreateMessageStatus(ctx context.Context, messageID uuid.UUID, recipientIDs []uuid.UUID) error
	UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, userID uuid.UUID, status MessageStatusType) error
	GetUnreadMessagesCount(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) (int, error)
	MarkMessagesAsRead(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) error
	DeleteMessage(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error
}

// CreateMessage saves a new message to database
// Returns the created message with generated ID and timestamp
func (pg *PostgresMessageStore) CreateMessage(ctx context.Context, msg *Message) (*Message, error) {
	msg.ID = uuid.New()

	query := `
        INSERT INTO messages (
            id, conversation_id, sender_id, content, 
            message_type, media_url, media_size, media_mime_type,
            reply_to_message_id
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING created_at
    `

	err := pg.DB.QueryRowContext(ctx, query,
		msg.ID,
		msg.ConversationID,
		msg.SenderID,
		msg.Content,
		msg.MessageType,
		msg.MediaURL,
		msg.MediaSize,
		msg.MediaMimeType,
		msg.ReplyToMessageID,
	).Scan(&msg.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	return msg, nil
}

// GetMessagesByConversationID fetches messages with pagination
// limit: how many messages to fetch
// before: fetch messages before this timestamp (for loading older messages)
func (pg *PostgresMessageStore) GetMessagesByConversationID(ctx context.Context, conversationID uuid.UUID, limit int, before *time.Time) ([]Message, error) {
	query := `
        SELECT 
            id, conversation_id, sender_id, content,
            message_type, media_url, media_size, media_mime_type,
            reply_to_message_id, created_at, edited_at, deleted_at
        FROM messages
        WHERE conversation_id = $1 
            AND deleted_at IS NULL
    `
	args := []interface{}{conversationID}
	if before != nil {
		query += " AND created_at < $2"
		args = append(args, before)
	}
	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)
	rows, err := pg.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []Message
	rows.Next()
	{
		var msg Message
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.Content,
			&msg.MessageType,
			&msg.MediaURL,
			&msg.MediaSize,
			&msg.MediaMimeType,
			&msg.ReplyToMessageID,
			&msg.CreatedAt,
			&msg.EditedAt,
			&msg.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)

	}

	return messages, nil
}

// CreateMessageStatus creates initial status records for all recipients
// Called immediately after creating a message
func (pg *PostgresMessageStore) CreateMessageStatus(ctx context.Context, messageID uuid.UUID, recipientIDs []uuid.UUID) error {
	tx, err := pg.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil
	}
	defer tx.Rollback()
	query := `
	INSERT INTO message_status (message_id,user_id,status)
	VALUES ($1,$2,$3)
	`
	for _, recipientID := range recipientIDs {
		_, err = tx.ExecContext(ctx, query, messageID, recipientID, MessageStatusSent)
		if err != nil {
			return nil
		}
	}
	return tx.Commit()
}

// UpdateMessageStatus updates delivery/read status for a user

func (pg *PostgresMessageStore) UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, userID uuid.UUID, status MessageStatusType) error {
	query := `
	UPDATE message_status 
	SET status = $1, timestamp = NOW()
	WHERE message_id = $2 AND user_id = $3
	`
	_, err := pg.DB.ExecContext(ctx, query, status, messageID, userID)
	return err
}

func (pg *PostgresMessageStore) GetUnreadMessagesCount(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) (int, error) {
	query := `
	SELECT COUNT(*) 
	FROM message_status ms
	INNER JOIN messages m ON ms.message_id = m.id
	WHERE m.user_id = $1
	AND m.conversation_id = $2
	AND ms.status != 'read'
	AND m.deleted_at IS NULL
	`
	var count int
	err := pg.DB.QueryRowContext(ctx, query, userID, conversationID).Scan(&count)

	return count, err
}

// MarkMessagesAsRead marks all messages in a conversation as read for a user
// Called when user opens a conversation
func (pg *PostgresMessageStore) MarkMessagesAsRead(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) error {
	query := `
	UPDATE message_status 
	SET status = 'read', timestamp = NOW()
	WHERE user_id = $1
	AND message_id IN (
	SELECT id FROM messages WHERE conversation_id = $2
	)
	AND status != 'read'
	`
	_, err := pg.DB.ExecContext(ctx, query, userID, conversationID)
	return err
}

func (pg *PostgresMessageStore) DeleteMessage(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error {
	query := `
        UPDATE messages
        SET deleted_at = NOW()
        WHERE id = $1 AND sender_id = $2 AND deleted_at IS NULL
    `

	result, err := pg.DB.ExecContext(ctx, query, messageID, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("message not found or already deleted")
	}

	return nil
}
