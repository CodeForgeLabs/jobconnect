-- Revert strict portfolio media source constraints.

ALTER TABLE portfolio_media
    DROP CONSTRAINT IF EXISTS chk_portfolio_media_source;
