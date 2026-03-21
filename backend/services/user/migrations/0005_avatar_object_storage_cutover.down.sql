ALTER TABLE profile_avatars
    ADD COLUMN IF NOT EXISTS content BYTEA;

UPDATE profile_avatars
SET content = ''::bytea
WHERE content IS NULL;

ALTER TABLE profile_avatars
    ALTER COLUMN content SET NOT NULL;

ALTER TABLE profile_avatars
    DROP COLUMN IF EXISTS storage_key;
