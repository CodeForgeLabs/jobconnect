-- Remove deprecated freelancer experience_level column from legacy schema variants.
ALTER TABLE freelancer_profiles
    DROP COLUMN IF EXISTS experience_level;
