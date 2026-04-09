-- Enforce contract-aligned portfolio constraints.

ALTER TABLE portfolio_items
    ALTER COLUMN title TYPE VARCHAR(50);

ALTER TABLE portfolio_items
    DROP CONSTRAINT IF EXISTS chk_portfolio_items_description_len,
    ADD CONSTRAINT chk_portfolio_items_description_len
        CHECK (description IS NULL OR char_length(description) <= 200);

ALTER TABLE portfolio_media
    DROP CONSTRAINT IF EXISTS chk_portfolio_media_source;

ALTER TABLE portfolio_media
    ADD CONSTRAINT chk_portfolio_media_source
        CHECK (
            (
                media_type = 'LINK'
                AND NULLIF(BTRIM(external_url), '') IS NOT NULL
                AND NULLIF(BTRIM(COALESCE(storage_key, '')), '') IS NULL
            )
            OR
            (
                media_type IN ('IMAGE', 'VIDEO', 'FILE')
                AND NULLIF(BTRIM(storage_key), '') IS NOT NULL
                AND NULLIF(BTRIM(COALESCE(external_url, '')), '') IS NULL
            )
        ) NOT VALID;
