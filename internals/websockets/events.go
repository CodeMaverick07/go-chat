package websockets

import (
	"context"
	"encoding/json"
	"fmt"
	"go-chat/internals/store"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, client *Client) error

const (
	// Outgoing events (server -> client)
	EventNewMessage          = "new_message"
	EventMessageDelivered    = "message_delivered"
	EventMessageRead         = "message_read"
	EventConversationCreated = "conversation_created"
	EventTypingIndicator     = "typing_indicator"
	// Incoming events (client -> server)
	EventSendMessage        = "send_message"
	EventCreateConversation = "create_conversation"
	EventGetConversations   = "get_conversations"
	EventGetMessages        = "get_messages"
	EventMarkAsRead         = "mark_as_read"
	EventStartTyping        = "start_typing"
	EventStopTyping         = "stop_typing"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}
type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}
type ChangeRoomEvent struct {
	Name string `json:"name"`
}

func ChatRoomHandler(event Event, c *Client) error {
	var changeRoomEvent ChangeRoomEvent
	if err := json.Unmarshal(event.Payload, &changeRoomEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}
	c.chatroom = changeRoomEvent.Name

	return nil

}

// SendMessagePayload is sent by client to send a message
// client ---> server
type SendMessagePayload struct {
	ConversationID   *uuid.UUID        `json:"conversation_id,omitempty"` // For group chat
	RecipientID      *uuid.UUID        `json:"recipient_id,omitempty"`    // For direct chat
	Content          string            `json:"content"`
	MessageType      store.MessageType `json:"message_type"`
	ReplyToMessageID *uuid.UUID        `json:"reply_to_message_id,omitempty"`
}

// NewMessagePayload is sent to all participants when a new message arrives
// server ---> client
type NewMessagePayload struct {
	MessageID      uuid.UUID         `json:"message_id"`
	ConversationID uuid.UUID         `json:"conversation_id"`
	SenderID       uuid.UUID         `json:"sender_id"`
	Content        string            `json:"content"`
	MessageType    store.MessageType `json:"message_type"`
	MediaURL       *string           `json:"media_url,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

// SendMessageHandler handles incoming send_message events

func SendMessageHandler(event Event, c *Client) error {
	var payload SendMessagePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	ctx := context.Background()

	var (
		message    *store.Message
		recipients []uuid.UUID
		err        error
	)

	switch {
	case payload.RecipientID != nil:
		message, recipients, err = c.Manager.MessageHandler.SendDirectMessage(
			ctx,
			c.UserID,
			*payload.RecipientID,
			payload.Content,
			payload.MessageType,
			nil, // mediaURL
			nil, // mediaSize
			nil, // mediaMimeType
		)

	case payload.ConversationID != nil:
		message, recipients, err = c.Manager.MessageHandler.SendGroupMessage(
			ctx,
			c.UserID,
			*payload.ConversationID,
			payload.Content,
			payload.MessageType,
			nil, // mediaURL
			nil, // mediaSize
			nil, // mediaMimeType
		)

	default:
		return fmt.Errorf("either recipient_id or conversation_id must be provided")
	}

	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	if message == nil {
		return fmt.Errorf("message is nil after send")
	}

	broadcastPayload := NewMessagePayload{
		MessageID:      message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		Content:        message.Content,
		MessageType:    message.MessageType,
		MediaURL:       message.MediaURL,
		CreatedAt:      message.CreatedAt,
	}

	data, err := json.Marshal(broadcastPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast payload: %w", err)
	}

	outgoingEvent := Event{
		Type:    EventNewMessage,
		Payload: data,
	}

	c.Manager.BroadcastToUsers(outgoingEvent, recipients)

	return nil
}

type CreateConversationPayload struct {
	Type           store.ConversationType `json:"type"`
	Name           *string                `json:"name,omitempty"`
	ParticipantIDs []uuid.UUID            `json:"participant_ids"`
}

type ConversationCreatedPayload struct {
	ConversationID uuid.UUID              `json:"conversation_id"`
	Type           store.ConversationType `json:"type"`
	Name           *string                `json:"name,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

func CreateConversationHandler(event Event, c *Client) error {
	var payload CreateConversationPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}
	ctx := context.Background()
	if payload.Type != store.ConversationTypeGroup {
		return fmt.Errorf("use send_message for direct chats")
	}

	if payload.Name == nil || *payload.Name == "" {
		return fmt.Errorf("group name is required")
	}
	conv, err := c.Manager.ConversationHandler.CreateGroupConversation(
		ctx,
		*payload.Name,
		c.UserID,
		payload.ParticipantIDs,
	)
	if err != nil {
		return err
	}
	responsePayload := ConversationCreatedPayload{
		ConversationID: conv.ID,
		Type:           conv.Type,
		Name:           conv.Name,
		CreatedAt:      conv.CreatedAt,
	}
	data, _ := json.Marshal(responsePayload)
	outgoingEvent := Event{
		Type:    EventConversationCreated,
		Payload: data,
	}
	allParticipants := append(payload.ParticipantIDs, c.UserID)
	c.Manager.BroadcastToUsers(outgoingEvent, allParticipants)
	return nil

}

type GetConversationsPayload struct {
	// Empty for now, could add filters later
}

type ConversationsListPayload struct {
	Conversations []store.ConversationWithDetails `json:"conversations"`
}

func GetConversationHandler(e Event, c *Client) error {
	ctx := context.Background()
	conversations, err := c.Manager.ConversationHandler.GetUserConversations(ctx, c.UserID)
	if err != nil {
		return err
	}
	responsePayload := ConversationsListPayload{
		Conversations: conversations,
	}
	data, _ := json.Marshal(responsePayload)
	c.egress <- Event{
		Type:    "conversations_list",
		Payload: data,
	}
	return nil
}

type GetMessagesPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	Limit          int       `json:"limit"`
	Before         *string   `json:"before,omitempty"`
}
type MessagesListPayload struct {
	ConversationID uuid.UUID       `json:"conversation_id"`
	Messages       []store.Message `json:"messages"`
}

func GetMessagesHandler(e Event, c *Client) error {
	var payload GetMessagesPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return err
	}
	ctx := context.Background()
	if payload.Limit == 0 {
		payload.Limit = 50
	}
	var before *time.Time
	if payload.Before != nil {
		t, err := time.Parse(time.RFC1123, *payload.Before)
		if err == nil {
			before = &t
		}
	}
	messages, err := c.Manager.ConversationHandler.GetConversationMessages(ctx, payload.ConversationID, c.UserID, payload.Limit, before)
	if err != nil {
		return err
	}
	responsePayload := MessagesListPayload{
		ConversationID: payload.ConversationID,
		Messages:       messages,
	}
	data, _ := json.Marshal(responsePayload)
	c.egress <- Event{
		Type:    "messages_list",
		Payload: data,
	}
	return nil
}

type MarkAsReadPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
}

func MarkAsReadHandler(event Event, c *Client) error {
	var payload MarkAsReadPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	ctx := context.Background()

	err := c.Manager.ConversationHandler.MarkConversationAsRead(ctx, c.UserID, payload.ConversationID)
	if err != nil {
		return err
	}

	// Optionally notify other participants (for read receipts)
	// This would show blue checkmarks in the sender's UI

	return nil
}

type TypingPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	IsTyping       bool      `json:"is_typing"`
}

func TypingIndicatorHandler(e Event, c *Client) error {
	var payload TypingPayload
	err := json.Unmarshal(e.Payload, &payload)
	if err != nil {
		return err
	}
	ctx := context.Background()
	participants, err := c.Manager.ConversationHandler.GetConversationParticipants(ctx, payload.ConversationID)
	if err != nil {
		return err
	}
	var recipients []uuid.UUID
	for _, participant := range participants {
		if participant != c.UserID {
			recipients = append(recipients, participant)
		}
	}
	data, _ := json.Marshal(map[string]interface{}{
		"conversation_id": payload.ConversationID,
		"user_id":         c.UserID,
		"is_typing":       payload.IsTyping,
	})
	c.Manager.BroadcastToUsers(Event{
		Type:    EventTypingIndicator,
		Payload: data,
	}, recipients)
	return nil

}
