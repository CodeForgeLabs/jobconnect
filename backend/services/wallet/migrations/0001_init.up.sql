CREATE TABLE wallet_accounts (
    id BIGSERIAL PRIMARY KEY,
    owner_id UUID UNIQUE NOT NULL,
    balance_minor BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE wallet_transactions (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NOT NULL REFERENCES wallet_accounts(id) ON DELETE RESTRICT,
    tx_ref VARCHAR(128) UNIQUE NOT NULL,
    chapa_ref VARCHAR(128),
    amount_minor BIGINT NOT NULL,
    tx_type VARCHAR(32) NOT NULL, -- e.g., 'deposit', 'withdrawal', 'payment'
    description TEXT NOT NULL,      -- e.g., 'Internal transfer to @user123'
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW()
);