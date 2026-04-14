-- Restore legacy verification_status columns on extension tables for rollback compatibility.
ALTER TABLE freelancer_profiles
    ADD COLUMN IF NOT EXISTS verification_status VARCHAR;

ALTER TABLE client_profiles
    ADD COLUMN IF NOT EXISTS verification_status VARCHAR;
