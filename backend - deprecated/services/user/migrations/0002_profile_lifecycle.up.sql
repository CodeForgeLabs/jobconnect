ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS language VARCHAR,
    ADD COLUMN IF NOT EXISTS contact_email VARCHAR,
    ADD COLUMN IF NOT EXISTS contact_phone VARCHAR,
    ADD COLUMN IF NOT EXISTS bio TEXT,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

ALTER TABLE client_profiles
    ADD CONSTRAINT client_profiles_profile_id_unique UNIQUE (profile_id);

ALTER TABLE freelancer_profiles
    ADD CONSTRAINT freelancer_profiles_profile_id_unique UNIQUE (profile_id);

CREATE TABLE IF NOT EXISTS profile_avatars (
    user_id UUID PRIMARY KEY REFERENCES profiles(user_id) ON DELETE CASCADE,
    file_name VARCHAR NOT NULL,
    content_type VARCHAR NOT NULL,
    content BYTEA NOT NULL,
    width INT NOT NULL,
    height INT NOT NULL,
    size_bytes BIGINT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
