ALTER TABLE wallet_accounts
    DROP CONSTRAINT IF EXISTS wallet_accounts_owner_currency_uniq,
    DROP COLUMN IF EXISTS currency;
