package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	FindOrCreateDirectConversation(ctx context.Context, user1ID uuid.UUID, user2ID uuid.UUID) (*Conversation, error)
	CreateGroupConversation(ctx context.Context, name string, creatorID uuid.UUID, participantIDs []uuid.UUID) (*Conversation, error)
	GetConversationsByUserID(ctx context.Context, userID uuid.UUID) ([]ConversationWithDetails, error)
	GetConversationParticipants(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error)
}

// FindOrCreateDirectConversation finds existing 1-1 chat or creates new one
func (pg *PostgresConversationStore) FindOrCreateDirectConversation(ctx context.Context, user1ID uuid.UUID, user2ID uuid.UUID) (*Conversation, error) {
	query := `
        SELECT DISTINCT c.id, c.type, c.name, c.created_by, c.created_at, c.updated_at
        FROM conversations c
        INNER JOIN conversation_participants cp1 ON c.id = cp1.conversation_id
        INNER JOIN conversation_participants cp2 ON c.id = cp2.conversation_id
        WHERE c.type = 'direct'
            AND cp1.user_id = $1
            AND cp2.user_id = $2
            AND cp1.left_at IS NULL
            AND cp2.left_at IS NULL
        LIMIT 1
    `
	var conv Conversation
	err := pg.DB.QueryRowContext(ctx, query, user1ID, user2ID).Scan(
		&conv.ID,
		&conv.Type,
		&conv.Name,
		&conv.CreatedBy,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)
	if err == nil {
		return &conv, err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("error finding conversation: %w", err)
	}
	tx, err := pg.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()
	conversationID := uuid.New()
	insertConv := `
        INSERT INTO conversations (id, type, created_by)
        VALUES ($1, $2, $3)
        RETURNING created_at, updated_at
    `
	err = tx.QueryRowContext(ctx, insertConv, conversationID, ConversationTypeDirect, user1ID).Scan(&conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error creating conversation: %w", err)
	}
	conv.ID = conversationID
	conv.Type = ConversationTypeDirect
	conv.CreatedBy = &user1ID
	insertParticipant := `
        INSERT INTO conversation_participants (conversation_id, user_id, role)
        VALUES ($1, $2, $3)
    `
	// User 1
	_, err = tx.ExecContext(ctx, insertParticipant, conversationID, user1ID, "member")
	if err != nil {
		return nil, fmt.Errorf("error adding participant 1: %w", err)
	}
	// User 2
	_, err = tx.ExecContext(ctx, insertParticipant, conversationID, user2ID, "member")
	if err != nil {
		return nil, fmt.Errorf("error adding participant 2: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &conv, nil
}
func (pg *PostgresConversationStore) CreateGroupConversation(
	ctx context.Context, name string, creatorID uuid.UUID, participantIDs []uuid.UUID) (
	*Conversation, error) {
	if len(participantIDs) < 1 {
		return nil, errors.New("group must have at least 2 participants")
	}
	tx, err := pg.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	conversationID := uuid.New()
	var conv Conversation
	query := `
        INSERT INTO conversations (id, type, name, created_by)
        VALUES ($1, $2, $3, $4)
        RETURNING id, type, name, created_by, created_at, updated_at
    `
	err = tx.QueryRowContext(ctx, query, conversationID, ConversationTypeGroup, name, creatorID).
		Scan(&conv.ID, &conv.Type, &conv.Name, &conv.CreatedBy, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil, err
	}
	insertParticipant := `
        INSERT INTO conversation_participants (conversation_id, user_id, role)
        VALUES ($1, $2, $3)
    `
	_, err = tx.ExecContext(ctx, insertParticipant, conversationID, creatorID, "admin")
	if err != nil {
		return nil, err
	}
	for _, participantID := range participantIDs {
		if participantID == creatorID {
			continue
		}
		_, err = tx.ExecContext(ctx, insertParticipant, conversationID, participantID, "member")
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &conv, nil
}
func (pg *PostgresConversationStore) GetConversationsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]ConversationWithDetails, error) {

	// Complex query joining multiple tables
	// Gets conversation info + last message + unread count
	query := `
        SELECT 
            c.id, c.type, c.name, c.created_by, c.created_at, c.updated_at,
            COUNT(DISTINCT cp.user_id) as participant_count,
            m.id as last_message_id,
            m.content as last_message_content,
            m.message_type as last_message_type,
            m.created_at as last_message_time,
            COUNT(CASE WHEN ms.status != 'read' AND ms.user_id = $1 THEN 1 END) as unread_count
        FROM conversations c
        INNER JOIN conversation_participants my_participation 
            ON c.id = my_participation.conversation_id 
            AND my_participation.user_id = $1
            AND my_participation.left_at IS NULL
        LEFT JOIN conversation_participants cp 
            ON c.id = cp.conversation_id 
            AND cp.left_at IS NULL
        LEFT JOIN LATERAL (
            SELECT id, content, message_type, created_at
            FROM messages
            WHERE conversation_id = c.id AND deleted_at IS NULL
            ORDER BY created_at DESC
            LIMIT 1
        ) m ON true
        LEFT JOIN message_status ms 
            ON m.id = ms.message_id 
            AND ms.user_id = $1
        GROUP BY c.id, c.type, c.name, c.created_by, c.created_at, c.updated_at, 
                 m.id, m.content, m.message_type, m.created_at
        ORDER BY m.created_at DESC NULLS LAST
    `

	rows, err := pg.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []ConversationWithDetails
	for rows.Next() {
		var conv ConversationWithDetails
		var lastMsgID, lastMsgContent, lastMsgType sql.NullString
		var lastMsgTime sql.NullTime

		err := rows.Scan(
			&conv.ID, &conv.Type, &conv.Name, &conv.CreatedBy,
			&conv.CreatedAt, &conv.UpdatedAt,
			&conv.ParticipantCount,
			&lastMsgID, &lastMsgContent, &lastMsgType, &lastMsgTime,
			&conv.UnreadCount,
		)
		if err != nil {
			return nil, err
		}
		if lastMsgID.Valid {
			msgID, _ := uuid.Parse(lastMsgID.String)
			conv.LastMessage = &Message{
				ID:          msgID,
				Content:     lastMsgContent.String,
				MessageType: MessageType(lastMsgType.String),
				CreatedAt:   lastMsgTime.Time,
			}
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}
func (pg *PostgresConversationStore) GetConversationParticipants(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error) {
	query := `
        SELECT user_id
        FROM conversation_participants
        WHERE conversation_id = $1 AND left_at IS NULL
    `
	rows, err := pg.DB.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var participants []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		participants = append(participants, userID)
	}
	return participants, nil
}
