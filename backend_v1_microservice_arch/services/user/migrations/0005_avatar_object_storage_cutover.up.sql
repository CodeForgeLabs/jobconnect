ALTER TABLE profile_avatars
    ADD COLUMN IF NOT EXISTS storage_key VARCHAR;

-- Hard cutover: remove legacy blob-backed avatar rows.
DELETE FROM profile_avatars;

ALTER TABLE profile_avatars
    DROP COLUMN IF EXISTS content;

ALTER TABLE profile_avatars
    ALTER COLUMN storage_key SET NOT NULL;
