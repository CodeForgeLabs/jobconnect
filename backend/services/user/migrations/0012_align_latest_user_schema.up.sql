-- Align user-service schema to latest contract/table specification.

-- 1) profiles: add tax_id + verification_status at profile level.
ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS tax_id VARCHAR,
    ADD COLUMN IF NOT EXISTS verification_status VARCHAR;

-- Backfill from legacy extension tables before dropping legacy columns.
UPDATE profiles p
SET tax_id = cp.tax_id
FROM client_profiles cp
WHERE cp.profile_id = p.id
  AND p.tax_id IS NULL
  AND cp.tax_id IS NOT NULL;

UPDATE profiles p
SET verification_status = fp.verification_status
FROM freelancer_profiles fp
WHERE fp.profile_id = p.id
  AND p.verification_status IS NULL
  AND fp.verification_status IS NOT NULL;

UPDATE profiles p
SET verification_status = cp.verification_status
FROM client_profiles cp
WHERE cp.profile_id = p.id
  AND p.verification_status IS NULL
  AND cp.verification_status IS NOT NULL;

-- 2) client_profiles: keep only latest-table fields.
ALTER TABLE client_profiles
    DROP COLUMN IF EXISTS tax_id,
    DROP COLUMN IF EXISTS verification_status,
    DROP COLUMN IF EXISTS created_at;

-- 3) freelancer_profiles: remove legacy fields and normalize earnings column name.
ALTER TABLE freelancer_profiles
    DROP COLUMN IF EXISTS bio;

ALTER TABLE freelancer_profiles
    RENAME COLUMN total_earnings_usd TO total_earnings;

-- 4) portfolio_items: enforce latest title/description limits.
UPDATE portfolio_items
SET title = LEFT(title, 50)
WHERE char_length(title) > 50;

ALTER TABLE portfolio_items
    ALTER COLUMN title TYPE VARCHAR(50);

ALTER TABLE portfolio_items
    DROP CONSTRAINT IF EXISTS chk_portfolio_items_description_len,
    ADD CONSTRAINT chk_portfolio_items_description_len
        CHECK (description IS NULL OR char_length(description) <= 200);

-- 5) portfolio_media: drop sort_order and simplify index.
DROP INDEX IF EXISTS idx_portfolio_media_item_sort;

ALTER TABLE portfolio_media
    DROP COLUMN IF EXISTS sort_order;

CREATE INDEX IF NOT EXISTS idx_portfolio_media_item_id
    ON portfolio_media(portfolio_item_id);

-- 6) freelancer_work_preferences: rename min/max budget columns.
ALTER TABLE freelancer_work_preferences
    RENAME COLUMN min_budget_usd TO min_budget;

ALTER TABLE freelancer_work_preferences
    RENAME COLUMN max_budget_usd TO max_budget;

-- 7) client_freelancer_notes: enforce max 100 chars for note.
ALTER TABLE client_freelancer_notes
    DROP CONSTRAINT IF EXISTS chk_client_freelancer_notes_note_len,
    ADD CONSTRAINT chk_client_freelancer_notes_note_len
        CHECK (char_length(note) <= 100);

-- 8) remove legacy impersonation-token table from user-service schema scope.
DROP INDEX IF EXISTS idx_admin_impersonation_tokens_target;
DROP TABLE IF EXISTS admin_impersonation_tokens;
