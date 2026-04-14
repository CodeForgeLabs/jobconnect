ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS location VARCHAR;

UPDATE profiles p
SET location = fp.location
FROM freelancer_profiles fp
WHERE fp.profile_id = p.id
  AND (p.location IS NULL OR p.location = '')
  AND fp.location IS NOT NULL
  AND fp.location <> '';

ALTER TABLE freelancer_profiles
    DROP COLUMN IF EXISTS location;
