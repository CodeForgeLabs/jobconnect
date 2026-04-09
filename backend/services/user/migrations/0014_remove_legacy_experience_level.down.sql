-- Restore deprecated freelancer experience_level column for rollback compatibility.
ALTER TABLE freelancer_profiles
    ADD COLUMN IF NOT EXISTS experience_level VARCHAR;
