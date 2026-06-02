DROP TABLE IF EXISTS profile_avatars;

ALTER TABLE freelancer_profiles
    DROP CONSTRAINT IF EXISTS freelancer_profiles_profile_id_unique;

ALTER TABLE client_profiles
    DROP CONSTRAINT IF EXISTS client_profiles_profile_id_unique;

ALTER TABLE profiles
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS bio,
    DROP COLUMN IF EXISTS contact_phone,
    DROP COLUMN IF EXISTS contact_email,
    DROP COLUMN IF EXISTS language;
