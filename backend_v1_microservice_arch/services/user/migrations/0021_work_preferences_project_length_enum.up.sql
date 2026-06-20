-- Normalize existing values and enforce canonical project-length tokens.
UPDATE freelancer_work_preferences
SET preferred_project_length = CASE
    WHEN trim(coalesce(preferred_project_length, '')) = '' THEN 'PROJECT_LENGTH_UNSPECIFIED'
    WHEN upper(trim(preferred_project_length)) IN ('UNSPECIFIED', 'PROJECT_LENGTH_UNSPECIFIED', 'NO_PREFERENCE', 'NONE') THEN 'PROJECT_LENGTH_UNSPECIFIED'
    WHEN upper(trim(preferred_project_length)) IN ('SHORT', 'SHORT_TERM', 'SHORT-TERM', 'PROJECT_LENGTH_SHORT_TERM') THEN 'PROJECT_LENGTH_SHORT_TERM'
    WHEN upper(trim(preferred_project_length)) IN ('MEDIUM', 'MEDIUM_TERM', 'MEDIUM-TERM', 'PROJECT_LENGTH_MEDIUM_TERM') THEN 'PROJECT_LENGTH_MEDIUM_TERM'
    WHEN upper(trim(preferred_project_length)) IN ('LONG', 'LONG_TERM', 'LONG-TERM', 'PROJECT_LENGTH_LONG_TERM') THEN 'PROJECT_LENGTH_LONG_TERM'
    ELSE 'PROJECT_LENGTH_UNSPECIFIED'
END;

ALTER TABLE freelancer_work_preferences
    ALTER COLUMN preferred_project_length SET DEFAULT 'PROJECT_LENGTH_UNSPECIFIED';

ALTER TABLE freelancer_work_preferences
    DROP CONSTRAINT IF EXISTS freelancer_work_preferences_project_length_check;

ALTER TABLE freelancer_work_preferences
    ADD CONSTRAINT freelancer_work_preferences_project_length_check
    CHECK (preferred_project_length IN (
        'PROJECT_LENGTH_UNSPECIFIED',
        'PROJECT_LENGTH_SHORT_TERM',
        'PROJECT_LENGTH_MEDIUM_TERM',
        'PROJECT_LENGTH_LONG_TERM'
    ));
