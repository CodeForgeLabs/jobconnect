CREATE TABLE IF NOT EXISTS wallet_accounts (
    id BIGSERIAL PRIMARY KEY,
    owner_id UUID NOT NULL,
    currency VARCHAR(8) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    available_minor BIGINT NOT NULL DEFAULT 0,
    held_minor BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT wallet_accounts_owner_currency_uniq UNIQUE (owner_id, currency),
    CONSTRAINT wallet_accounts_available_nonnegative CHECK (available_minor >= 0),
    CONSTRAINT wallet_accounts_held_nonnegative CHECK (held_minor >= 0)
);

CREATE TABLE IF NOT EXISTS wallet_holds (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NOT NULL REFERENCES wallet_accounts(id) ON DELETE CASCADE,
    reference_type VARCHAR(64) NOT NULL,
    reference_id VARCHAR(128) NOT NULL,
    amount_minor BIGINT NOT NULL,
    captured_minor BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL,
    expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    released_at TIMESTAMPTZ NULL,
    captured_at TIMESTAMPTZ NULL,
    CONSTRAINT wallet_holds_wallet_reference_uniq UNIQUE (wallet_id, reference_type, reference_id),
    CONSTRAINT wallet_holds_amount_positive CHECK (amount_minor > 0),
    CONSTRAINT wallet_holds_captured_nonnegative CHECK (captured_minor >= 0),
    CONSTRAINT wallet_holds_captured_not_exceed_amount CHECK (captured_minor <= amount_minor)
);

CREATE TABLE IF NOT EXISTS wallet_ledger_entries (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NOT NULL REFERENCES wallet_accounts(id) ON DELETE CASCADE,
    entry_type VARCHAR(32) NOT NULL,
    amount_minor BIGINT NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    reference_type VARCHAR(64) NOT NULL DEFAULT '',
    reference_id VARCHAR(128) NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    available_after_minor BIGINT NOT NULL,
    held_after_minor BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT wallet_ledger_entries_wallet_idempotency_uniq UNIQUE (wallet_id, idempotency_key),
    CONSTRAINT wallet_ledger_entries_amount_positive CHECK (amount_minor > 0),
    CONSTRAINT wallet_ledger_entries_available_nonnegative CHECK (available_after_minor >= 0),
    CONSTRAINT wallet_ledger_entries_held_nonnegative CHECK (held_after_minor >= 0)
);

CREATE INDEX IF NOT EXISTS wallet_holds_wallet_id_idx ON wallet_holds (wallet_id);
CREATE INDEX IF NOT EXISTS wallet_holds_status_idx ON wallet_holds (status);
CREATE INDEX IF NOT EXISTS wallet_ledger_entries_wallet_created_idx ON wallet_ledger_entries (wallet_id, created_at DESC);
