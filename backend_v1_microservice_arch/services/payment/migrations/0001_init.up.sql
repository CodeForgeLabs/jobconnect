CREATE TABLE IF NOT EXISTS payment_sessions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL,
    provider        VARCHAR(32) NOT NULL,
    payment_type    VARCHAR(32) NOT NULL,
    status          VARCHAR(32) NOT NULL DEFAULT 'pending',
    amount_minor    BIGINT NOT NULL,
    currency        VARCHAR(8) NOT NULL DEFAULT 'ETB',
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    external_ref    VARCHAR(256),
    receipt_key     VARCHAR(512),
    reference_type  VARCHAR(64),
    reference_id    VARCHAR(128),
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_payment_sessions_user ON payment_sessions (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_sessions_status ON payment_sessions (status);
CREATE INDEX IF NOT EXISTS idx_payment_sessions_external ON payment_sessions (external_ref);
