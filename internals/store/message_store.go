package store

import (
	"context"
	"database/sql"
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

func (pg *PostgresMessageStore) CreateMessage(ctx context.Context, msg *Message) (*Message, error) {
	return nil, nil
}

func (pg *PostgresMessageStore) GetMessagesByConversationID(ctx context.Context, conversationID uuid.UUID, limit int, before *time.Time) ([]Message, error) {
	return nil, nil
}

func (pg *PostgresMessageStore) CreateMessageStatus(ctx context.Context, messageID uuid.UUID, recipientIDs []uuid.UUID) error {
	return nil
}

func (pg *PostgresMessageStore) UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, userID uuid.UUID, status MessageStatusType) error {
	return nil
}

func (pg *PostgresMessageStore) GetUnreadMessagesCount(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) (int, error) {
	return 0, nil
}

func (pg *PostgresMessageStore) MarkMessagesAsRead(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) error {
	return nil
}

func (pg *PostgresMessageStore) DeleteMessage(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error {
	return nil
}
