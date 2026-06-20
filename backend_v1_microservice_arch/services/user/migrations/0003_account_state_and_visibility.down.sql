DROP INDEX IF EXISTS idx_profiles_visibility;
DROP INDEX IF EXISTS idx_profiles_account_status;

ALTER TABLE profiles
    DROP COLUMN IF EXISTS visibility,
    DROP COLUMN IF EXISTS suspension_reason,
    DROP COLUMN IF EXISTS account_status;
