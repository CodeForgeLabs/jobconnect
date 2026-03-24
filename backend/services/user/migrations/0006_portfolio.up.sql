CREATE TABLE IF NOT EXISTS portfolio_items (
    id BIGSERIAL PRIMARY KEY,
    profile_id INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    title VARCHAR(160) NOT NULL,
    description TEXT,
    project_url VARCHAR(2048),
    role_in_project VARCHAR(120),
    completed_at DATE,
    sort_order INT NOT NULL DEFAULT 0,
    visibility VARCHAR NOT NULL DEFAULT 'PUBLIC',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS portfolio_media (
    id BIGSERIAL PRIMARY KEY,
    portfolio_item_id BIGINT NOT NULL REFERENCES portfolio_items(id) ON DELETE CASCADE,
    media_type VARCHAR(32) NOT NULL,
    storage_key VARCHAR(512),
    external_url VARCHAR(2048),
    file_name VARCHAR(255),
    content_type VARCHAR(128),
    size_bytes BIGINT,
    width INT,
    height INT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS portfolio_tags (
    portfolio_item_id BIGINT NOT NULL REFERENCES portfolio_items(id) ON DELETE CASCADE,
    tag VARCHAR(64) NOT NULL,
    PRIMARY KEY (portfolio_item_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_portfolio_items_profile_visibility_sort
    ON portfolio_items(profile_id, visibility, sort_order, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_portfolio_media_item_sort
    ON portfolio_media(portfolio_item_id, sort_order);
