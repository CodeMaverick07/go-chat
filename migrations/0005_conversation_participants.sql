-- +goose Up
-- +goose StatementBegin
CREATE TABLE conversation_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- When user joined the conversation
    -- For groups: tracks when someone was added
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- When user left the conversation
    -- NULL means still active member
    -- For "leave group" functionality
    left_at TIMESTAMP WITH TIME ZONE,
    
    -- Role in the conversation
    -- 'admin': can add/remove members, change group name
    -- 'member': regular participant
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('admin', 'member')),
    
    -- Prevent duplicate entries
    -- Each user can only be in a conversation once
    UNIQUE(conversation_id, user_id)
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS conversation_participants;

-- +goose StatementEnd