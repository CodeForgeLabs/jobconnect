ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS account_status VARCHAR NOT NULL DEFAULT 'ACTIVE',
    ADD COLUMN IF NOT EXISTS suspension_reason TEXT,
    ADD COLUMN IF NOT EXISTS visibility VARCHAR NOT NULL DEFAULT 'PUBLIC';

CREATE INDEX IF NOT EXISTS idx_profiles_account_status ON profiles(account_status);
CREATE INDEX IF NOT EXISTS idx_profiles_visibility ON profiles(visibility);
