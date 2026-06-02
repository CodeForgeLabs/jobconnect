-- Remove portfolio sort ordering columns and keep deterministic indexes.

DROP INDEX IF EXISTS idx_portfolio_media_item_sort;

ALTER TABLE portfolio_items
    DROP COLUMN IF EXISTS sort_order;

ALTER TABLE portfolio_media
    DROP COLUMN IF EXISTS sort_order;

CREATE INDEX IF NOT EXISTS idx_portfolio_items_profile_created_id
    ON portfolio_items(profile_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_portfolio_media_item_id
    ON portfolio_media(portfolio_item_id, id ASC);
