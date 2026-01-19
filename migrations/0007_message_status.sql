-- +goose Up
-- +goose StatementBegin
CREATE TABLE message_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Status progression:
    -- 'sent': Message saved to database
    -- 'delivered': Message arrived at recipient's device
    -- 'read': Recipient opened the conversation and saw the message
    status VARCHAR(20) NOT NULL CHECK (status IN ('sent', 'delivered', 'read')),
    
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Each recipient has exactly one status record per message
    UNIQUE(message_id, user_id)
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS message_status;
-- +goose StatementEnd