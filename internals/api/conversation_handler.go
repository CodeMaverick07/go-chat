package api

import (
	"context"
	"errors"
	"go-chat/internals/store"
	"log"
	"time"

	"github.com/google/uuid"
)

type ConversationHandler struct {
	MessageStore      store.MessageStore
	ConversationStore store.ConversationStore
	Logger            *log.Logger
}

func NewConversationHandler(messageStore store.MessageStore, conversationStore store.ConversationStore, logger *log.Logger) *ConversationHandler {
	return &ConversationHandler{
		MessageStore:      messageStore,
		ConversationStore: conversationStore,
		Logger:            logger,
	}
}

func (c *ConversationHandler) CreateGroupConversation(
	ctx context.Context,
	name string,
	creatorID uuid.UUID,
	participantIDs []uuid.UUID,
) (*store.Conversation, error) {

	return c.ConversationStore.CreateGroupConversation(
		ctx, name, creatorID, participantIDs,
	)
}

// GetUserConversations fetches all conversations for a user
func (c *ConversationHandler) GetUserConversations(
	ctx context.Context,
	userID uuid.UUID,
) ([]store.ConversationWithDetails, error) {

	return c.ConversationStore.GetConversationsByUserID(ctx, userID)
}
func (c *ConversationHandler) GetConversationMessages(
	ctx context.Context,
	conversationID uuid.UUID,
	userID uuid.UUID,
	limit int,
	before *time.Time,
) ([]store.Message, error) {
	participants, err := c.ConversationStore.GetConversationParticipants(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	isMember := false
	for _, p := range participants {
		if p == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, errors.New("user is not a member of this conversation")
	}

	return c.MessageStore.GetMessagesByConversationID(ctx, conversationID, limit, before)
}

func (c *ConversationHandler) MarkConversationAsRead(
	ctx context.Context,
	userID uuid.UUID,
	conversationID uuid.UUID,
) error {

	return c.MessageStore.MarkMessagesAsRead(ctx, userID, conversationID)
}

func (c *ConversationHandler) GetConversationParticipants(
	ctx context.Context,
	conversationID uuid.UUID,
) ([]uuid.UUID, error) {
	return c.ConversationStore.GetConversationParticipants(ctx, conversationID)
}
