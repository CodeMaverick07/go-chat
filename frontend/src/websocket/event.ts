export const WS_EVENTS = {
  // (server -> client)
  EventNewMessage: "new_message",
  EventMessageDelivered: "message_delivered",
  EventMessageRead: "message_read",
  EventConversationCreated: "conversation_created",
  EventTypingIndicator: "typing_indicator",
  // (client -> server)
  EventSendMessage: "send_message",
  EventCreateConversation: "create_conversation",
  EventGetConversations: "get_conversations",
  EventGetMessages: "get_messages",
  EventMarkAsRead: "mark_as_read",
  EventStartTyping: "start_typing",
  EventStopTyping: "stop_typing",
};
