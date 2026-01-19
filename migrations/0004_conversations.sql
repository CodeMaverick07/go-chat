-- +goose Up
-- +goose StatementBegin
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Type: 'direct' for 1-1 chat, 'group' for group chat
    -- Direct: exactly 2 participants, no name needed
    -- Group: 2+ participants, has a name
    type VARCHAR(20) NOT NULL CHECK (type IN ('direct', 'group')),

    -- Group name (e.g., "Project Team", "Family")
    -- NULL for direct chats (we show other user's name instead)
    name VARCHAR(100),

    -- Who created this conversation
    -- Useful for group chats (original creator has special status)
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS conversations;
-- +goose StatementEnd