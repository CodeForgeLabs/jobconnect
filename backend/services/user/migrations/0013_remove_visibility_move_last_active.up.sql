-- Move last_active_at to profiles and remove visibility columns/indexes.

ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMP;

UPDATE profiles p
SET last_active_at = fp.last_active_at
FROM freelancer_profiles fp
WHERE fp.profile_id = p.id
  AND p.last_active_at IS NULL
  AND fp.last_active_at IS NOT NULL;

ALTER TABLE freelancer_profiles
    DROP COLUMN IF EXISTS last_active_at;

DROP INDEX IF EXISTS idx_profiles_visibility;
ALTER TABLE profiles
    DROP COLUMN IF EXISTS visibility;

DROP INDEX IF EXISTS idx_portfolio_items_profile_visibility_sort;
ALTER TABLE portfolio_items
    DROP COLUMN IF EXISTS visibility;

DROP INDEX IF EXISTS idx_profile_employment_profile_visibility_sort;
ALTER TABLE profile_employment
    DROP COLUMN IF EXISTS visibility;

DROP INDEX IF EXISTS idx_profile_education_profile_visibility_sort;
ALTER TABLE profile_education
    DROP COLUMN IF EXISTS visibility;

DROP INDEX IF EXISTS idx_profile_certifications_profile_visibility;
ALTER TABLE profile_certifications
    DROP COLUMN IF EXISTS visibility;

DROP INDEX IF EXISTS idx_profile_languages_profile_visibility;
ALTER TABLE profile_languages
    DROP COLUMN IF EXISTS visibility;
