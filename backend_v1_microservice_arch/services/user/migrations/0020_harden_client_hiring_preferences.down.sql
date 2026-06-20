ALTER TABLE client_hiring_preferences
    DROP CONSTRAINT IF EXISTS chk_client_hiring_preferences_min_lte_max,
    DROP CONSTRAINT IF EXISTS chk_client_hiring_preferences_non_negative_rates;

ALTER TABLE client_hiring_preferences
    ADD COLUMN IF NOT EXISTS preferred_experience_levels JSONB NOT NULL DEFAULT '[]'::jsonb;
