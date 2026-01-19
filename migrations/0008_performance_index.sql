-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_messages_conversation_time 
    ON messages(conversation_id, created_at DESC);

-- Find all conversations for a user
-- Used for conversation list screen
-- WHERE clause filters out users who left
CREATE INDEX idx_participants_user 
    ON conversation_participants(user_id) 
    WHERE left_at IS NULL;

-- Get all participants in a conversation
-- Used to know who to send messages to
CREATE INDEX idx_participants_conversation 
    ON conversation_participants(conversation_id) 
    WHERE left_at IS NULL;

-- Check unread messages for a user
-- For notification badges
CREATE INDEX idx_message_status_unread 
    ON message_status(user_id, status) 
    WHERE status != 'read';

-- Find direct conversation between two specific users
-- Prevents creating duplicate 1-1 chats
CREATE INDEX idx_participants_lookup 
    ON conversation_participants(user_id, conversation_id);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_messages_conversation_time;
DROP INDEX IF EXISTS idx_participants_user;
DROP INDEX IF EXISTS idx_participants_conversation;
DROP INDEX IF EXISTS idx_message_status_unread;
DROP INDEX IF EXISTS idx_participants_lookup;
-- +goose StatementEnd