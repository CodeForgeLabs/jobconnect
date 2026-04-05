-- Revert 0012_align_latest_user_schema.up.sql

-- 1) Restore legacy impersonation-token table.
CREATE TABLE IF NOT EXISTS admin_impersonation_tokens (
    token_id UUID PRIMARY KEY,
    admin_user_id UUID NOT NULL,
    target_user_id UUID NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_impersonation_tokens_target
    ON admin_impersonation_tokens(target_user_id, created_at DESC);

-- 2) Remove note-length constraint added in 0012.
ALTER TABLE client_freelancer_notes
    DROP CONSTRAINT IF EXISTS chk_client_freelancer_notes_note_len;

-- 3) Restore freelancer_work_preferences legacy column names.
ALTER TABLE freelancer_work_preferences
    RENAME COLUMN min_budget TO min_budget_usd;

ALTER TABLE freelancer_work_preferences
    RENAME COLUMN max_budget TO max_budget_usd;

-- 4) Restore portfolio_media sort_order and old index shape.
DROP INDEX IF EXISTS idx_portfolio_media_item_id;

ALTER TABLE portfolio_media
    ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_portfolio_media_item_sort
    ON portfolio_media(portfolio_item_id, sort_order);

-- 5) Revert portfolio_items title/description constraints.
ALTER TABLE portfolio_items
    DROP CONSTRAINT IF EXISTS chk_portfolio_items_description_len;

ALTER TABLE portfolio_items
    ALTER COLUMN title TYPE VARCHAR(160);

-- 6) Restore freelancer_profiles legacy fields and earnings column name.
ALTER TABLE freelancer_profiles
    RENAME COLUMN total_earnings TO total_earnings_usd;

ALTER TABLE freelancer_profiles
    ADD COLUMN IF NOT EXISTS bio TEXT,
    ADD COLUMN IF NOT EXISTS verification_status VARCHAR;

-- 7) Restore dropped legacy client_profiles columns.
ALTER TABLE client_profiles
    ADD COLUMN IF NOT EXISTS tax_id VARCHAR,
    ADD COLUMN IF NOT EXISTS verification_status VARCHAR,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT NOW();

-- 8) Drop profile-level columns introduced in 0012.
ALTER TABLE profiles
    DROP COLUMN IF EXISTS tax_id,
    DROP COLUMN IF EXISTS verification_status;
