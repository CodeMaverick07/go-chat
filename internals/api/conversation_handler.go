package api

import (
	"go-chat/internals/store"
	"log"
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
