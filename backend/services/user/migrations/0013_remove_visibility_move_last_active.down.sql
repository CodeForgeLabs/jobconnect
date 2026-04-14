-- Restore visibility columns/indexes and move last_active_at back to freelancer_profiles.

ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS visibility VARCHAR NOT NULL DEFAULT 'PUBLIC';

CREATE INDEX IF NOT EXISTS idx_profiles_visibility
    ON profiles(visibility);

ALTER TABLE portfolio_items
    ADD COLUMN IF NOT EXISTS visibility VARCHAR NOT NULL DEFAULT 'PUBLIC';

CREATE INDEX IF NOT EXISTS idx_portfolio_items_profile_visibility_sort
    ON portfolio_items(profile_id, visibility, sort_order, created_at DESC)
    WHERE deleted_at IS NULL;

ALTER TABLE profile_employment
    ADD COLUMN IF NOT EXISTS visibility VARCHAR NOT NULL DEFAULT 'PUBLIC';

CREATE INDEX IF NOT EXISTS idx_profile_employment_profile_visibility_sort
    ON profile_employment(profile_id, visibility, sort_order, start_date DESC)
    WHERE deleted_at IS NULL;

ALTER TABLE profile_education
    ADD COLUMN IF NOT EXISTS visibility VARCHAR NOT NULL DEFAULT 'PUBLIC';

CREATE INDEX IF NOT EXISTS idx_profile_education_profile_visibility_sort
    ON profile_education(profile_id, visibility, sort_order, start_date DESC)
    WHERE deleted_at IS NULL;

ALTER TABLE profile_certifications
    ADD COLUMN IF NOT EXISTS visibility VARCHAR NOT NULL DEFAULT 'PUBLIC';

CREATE INDEX IF NOT EXISTS idx_profile_certifications_profile_visibility
    ON profile_certifications(profile_id, visibility, issue_date DESC)
    WHERE deleted_at IS NULL;

ALTER TABLE profile_languages
    ADD COLUMN IF NOT EXISTS visibility VARCHAR NOT NULL DEFAULT 'PUBLIC';

CREATE INDEX IF NOT EXISTS idx_profile_languages_profile_visibility
    ON profile_languages(profile_id, visibility);

ALTER TABLE freelancer_profiles
    ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMP;

UPDATE freelancer_profiles fp
SET last_active_at = p.last_active_at
FROM profiles p
WHERE p.id = fp.profile_id
  AND fp.last_active_at IS NULL
  AND p.last_active_at IS NOT NULL;

ALTER TABLE profiles
    DROP COLUMN IF EXISTS last_active_at;
