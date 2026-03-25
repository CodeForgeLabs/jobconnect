-- enable uuid generation (run once in your DB)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- messages table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    sender_id UUID NOT NULL,
    receiver_id UUID NOT NULL,

    -- content type: text, image, video
    type VARCHAR(20) NOT NULL,

    -- content fields
    text TEXT,
    image_url TEXT,
    video_url TEXT,
    caption TEXT,

    -- seen status (1-to-1)
    is_seen BOOLEAN NOT NULL DEFAULT FALSE,
    seen_at TIMESTAMPTZ,

    -- edit status
    is_edited BOOLEAN NOT NULL DEFAULT FALSE,
    edited_at TIMESTAMPTZ,

    -- delete status (soft delete)
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- basic validation
    CONSTRAINT valid_message_content CHECK (
        (type = 'text' AND text IS NOT NULL AND text <> '') OR
        (type = 'image' AND image_url IS NOT NULL AND image_url <> '') OR
        (type = 'video' AND video_url IS NOT NULL AND video_url <> '')
    )
);

-- index for fast chat queries
CREATE INDEX idx_messages_conversation
ON messages(sender_id, receiver_id);

-- index for sorting
CREATE INDEX idx_messages_created_at
ON messages(created_at);