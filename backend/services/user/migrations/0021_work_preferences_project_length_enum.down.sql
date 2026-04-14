ALTER TABLE freelancer_work_preferences
    DROP CONSTRAINT IF EXISTS freelancer_work_preferences_project_length_check;

ALTER TABLE freelancer_work_preferences
    ALTER COLUMN preferred_project_length SET DEFAULT '';

UPDATE freelancer_work_preferences
SET preferred_project_length = CASE
    WHEN preferred_project_length = 'PROJECT_LENGTH_UNSPECIFIED' THEN ''
    WHEN preferred_project_length = 'PROJECT_LENGTH_SHORT_TERM' THEN 'short'
    WHEN preferred_project_length = 'PROJECT_LENGTH_MEDIUM_TERM' THEN 'medium'
    WHEN preferred_project_length = 'PROJECT_LENGTH_LONG_TERM' THEN 'long'
    ELSE coalesce(preferred_project_length, '')
END;
