ALTER TABLE freelancer_profiles
    ADD COLUMN IF NOT EXISTS location VARCHAR;

UPDATE freelancer_profiles fp
SET location = p.location
FROM profiles p
WHERE p.id = fp.profile_id
  AND (fp.location IS NULL OR fp.location = '')
  AND p.location IS NOT NULL
  AND p.location <> '';

ALTER TABLE profiles
    DROP COLUMN IF EXISTS location;
