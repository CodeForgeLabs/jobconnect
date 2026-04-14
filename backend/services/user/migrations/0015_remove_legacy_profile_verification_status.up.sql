-- Drop legacy verification_status columns kept on profile extension tables.
ALTER TABLE freelancer_profiles
    DROP COLUMN IF EXISTS verification_status;

ALTER TABLE client_profiles
    DROP COLUMN IF EXISTS verification_status;
