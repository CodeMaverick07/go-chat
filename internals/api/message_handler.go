package api

import (
	"context"
	"errors"
	"fmt"
	"go-chat/internals/store"
	"log"

	"github.com/google/uuid"
)

type MessageHandler struct {
	MessageStore      store.MessageStore
	ConversationStore store.ConversationStore
	Logger            *log.Logger
}

func NewMessageHandler(messageStore store.MessageStore, conversationStore store.ConversationStore, logger *log.Logger) *MessageHandler {
	return &MessageHandler{
		MessageStore:      messageStore,
		ConversationStore: conversationStore,
		Logger:            logger,
	}
}

func (m *MessageHandler) SendDirectMessage(
	ctx context.Context,
	senderID uuid.UUID,
	recipientID uuid.UUID,
	content string,
	messageType store.MessageType,
	mediaURL *string,
	mediaSize *int64,
	mediaMimeType *string) (
	*store.Message, []uuid.UUID, error,
) {
	conversation, err := m.ConversationStore.FindOrCreateDirectConversation(ctx, senderID, recipientID)
	if err != nil {
		return nil, nil, fmt.Errorf("error finding/creating conversation: %w", err)
	}
	msg := &store.Message{
		ConversationID: conversation.ID,
		SenderID:       senderID,
		Content:        content,
		MessageType:    messageType,
		MediaURL:       mediaURL,
		MediaSize:      mediaSize,
		MediaMimeType:  mediaMimeType,
	}
	msg, err = m.MessageStore.CreateMessage(ctx, msg)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating message: %w", err)
	}
	err = m.MessageStore.CreateMessageStatus(ctx, msg.ID, []uuid.UUID{recipientID})
	if err != nil {
		return nil, nil, fmt.Errorf("error creating message status: %w", err)
	}
	return msg, []uuid.UUID{senderID, recipientID}, nil
}

func (m *MessageHandler) SendGroupMessage(
	ctx context.Context,
	senderID uuid.UUID,
	conversationID uuid.UUID,
	content string,
	messageType store.MessageType,
	mediaURL *string,
	mediaSize *int64,
	mediaMimeType *string,
) (*store.Message, []uuid.UUID, error) {
	// Verify sender is a participant in this conversation
	participants, err := m.ConversationStore.GetConversationParticipants(ctx, conversationID)
	if err != nil {
		return nil, nil, err
	}
	isMember := false
	for _, p := range participants {
		if p == senderID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, nil, errors.New("sender is not a member of this conversation")
	}
	//Create the message
	msg := &store.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		MessageType:    messageType,
		MediaURL:       mediaURL,
		MediaSize:      mediaSize,
		MediaMimeType:  mediaMimeType,
	}
	msg, err = m.MessageStore.CreateMessage(ctx, msg)
	if err != nil {
		return nil, nil, err
	}
	//Create message status for all participants except sender
	var recipients []uuid.UUID
	for _, p := range participants {
		if p != senderID {
			recipients = append(recipients, p)
		}
	}
	err = m.MessageStore.CreateMessageStatus(ctx, msg.ID, recipients)
	if err != nil {
		return nil, nil, err
	}
	return msg, participants, nil
}

func (m *MessageHandler) UpdateMessageDeliveryStatus(
	ctx context.Context,
	messageID uuid.UUID,
	userID uuid.UUID,
) error {

	return m.MessageStore.UpdateMessageStatus(
		ctx, messageID, userID, store.MessageStatusDelivered,
	)
}
