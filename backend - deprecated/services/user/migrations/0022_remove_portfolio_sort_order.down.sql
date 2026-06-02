-- Restore portfolio sort ordering columns and indexes.

DROP INDEX IF EXISTS idx_portfolio_items_profile_created_id;

ALTER TABLE portfolio_items
    ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;

ALTER TABLE portfolio_media
    ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_portfolio_media_item_sort
    ON portfolio_media(portfolio_item_id, sort_order);

CREATE INDEX IF NOT EXISTS idx_portfolio_items_profile_sort_id
    ON portfolio_items(profile_id, sort_order, id ASC)
    WHERE deleted_at IS NULL;
