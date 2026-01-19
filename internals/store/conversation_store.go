package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ConversationType string

const (
	ConversationTypeDirect ConversationType = "direct"
	ConversationTypeGroup  ConversationType = "group"
)

type Conversation struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	Type      ConversationType `json:"type" db:"type"`
	Name      *string          `json:"name,omitempty" db:"name"`
	CreatedBy *uuid.UUID       `json:"created_by,omitempty" db:"created_by"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt time.Time        `json:"updated_at" db:"updated_at"`
}

type ConversationParticipant struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ConversationID uuid.UUID  `json:"conversation_id" db:"conversation_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	JoinedAt       time.Time  `json:"joined_at" db:"joined_at"`
	LeftAt         *time.Time `json:"left_at,omitempty" db:"left_at"`
	Role           string     `json:"role" db:"role"`
}

type PostgresConversationStore struct {
	DB *sql.DB
}

func NewPostgresConversationStore(db *sql.DB) *PostgresConversationStore {
	return &PostgresConversationStore{
		DB: db,
	}
}

type ConversationStore interface {
	FindOrCreateDirectConversation(ctx context.Context, user1ID uuid.UUID, user2ID uuid.UUID) *Conversation
	CreateGroupConversation(ctx context.Context, name string, creatorID uuid.UUID, participantIDs []uuid.UUID) (*Conversation, error)
	GetConversationsByUserID(ctx context.Context, userID uuid.UUID) (*ConversationWithDetails, error)
	GetConversationParticipants(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error)
}

func (pg *PostgresConversationStore) FindOrCreateDirectConversation(ctx context.Context, user1ID uuid.UUID, user2ID uuid.UUID) *Conversation {
	return nil
}
func (pg *PostgresConversationStore) CreateGroupConversation(ctx context.Context, name string, creatorID uuid.UUID, participantIDs []uuid.UUID) (*Conversation, error) {
	return nil, nil
}
func (pg *PostgresConversationStore) GetConversationsByUserID(ctx context.Context, userID uuid.UUID) (*ConversationWithDetails, error) {
	return nil, nil
}
func (pg *PostgresConversationStore) GetConversationParticipants(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
