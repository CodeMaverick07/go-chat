package api

import (
	"go-chat/internals/store"
	"log"
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
