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
