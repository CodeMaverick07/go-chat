-- +goose Up
-- +goose StatementBegin
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Which conversation this message belongs to
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    
    -- Who sent the message
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    
    -- Message content
    -- For text: the actual text
    -- For media: optional caption
    content TEXT,
    
    -- Message type determines how client displays it
    -- 'text': regular text message
    -- 'image': display as image with optional caption
    -- 'video': video player with optional caption
    -- 'file': download link with filename
    message_type VARCHAR(20) NOT NULL DEFAULT 'text' 
        CHECK (message_type IN ('text', 'image', 'video', 'file')),
    
    -- For media messages: where file is stored
    -- Example: "uploads/images/uuid-filename.jpg"
    -- Or S3 URL: "https://mybucket.s3.amazonaws.com/..."
    media_url TEXT,
    
    -- File size in bytes
    -- WHY: Show upload/download progress, warn if too large
    media_size BIGINT,
    
    -- MIME type of the file
    -- Examples: "image/jpeg", "video/mp4", "application/pdf"
    -- WHY: Client knows how to handle the file
    media_mime_type VARCHAR(100),
    
    -- Reply/Quote functionality
    -- If this message is replying to another message
    -- NULL for regular messages
    reply_to_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- If user edited their message
    -- WHY: Show "edited" badge in UI
    edited_at TIMESTAMP WITH TIME ZONE,
    
    -- Soft delete: message marked as deleted but not removed from DB
    -- WHY: Support "delete for me" vs "delete for everyone"
    -- Also preserves conversation flow even if messages deleted
    deleted_at TIMESTAMP WITH TIME ZONE
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS messages;
-- +goose StatementEnd