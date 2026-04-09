CREATE TABLE IF NOT EXISTS profile_cvs (
    user_id UUID PRIMARY KEY REFERENCES profiles(user_id) ON DELETE CASCADE,
    file_name VARCHAR NOT NULL,
    content_type VARCHAR NOT NULL,
    storage_key VARCHAR NOT NULL,
    size_bytes BIGINT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);