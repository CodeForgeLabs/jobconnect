ALTER TABLE wallet_accounts
    ADD COLUMN IF NOT EXISTS currency VARCHAR(8) NOT NULL DEFAULT 'ETB';

ALTER TABLE wallet_accounts
    ADD CONSTRAINT wallet_accounts_owner_currency_uniq UNIQUE (owner_id, currency);
