UPDATE client_hiring_preferences
SET
    min_hourly_rate = GREATEST(min_hourly_rate, 0),
    max_hourly_rate = GREATEST(max_hourly_rate, 0);

UPDATE client_hiring_preferences
SET min_hourly_rate = max_hourly_rate
WHERE max_hourly_rate > 0
  AND min_hourly_rate > max_hourly_rate;

ALTER TABLE client_hiring_preferences
    DROP COLUMN IF EXISTS preferred_experience_levels;

ALTER TABLE client_hiring_preferences
    ADD CONSTRAINT chk_client_hiring_preferences_non_negative_rates
        CHECK (min_hourly_rate >= 0 AND max_hourly_rate >= 0),
    ADD CONSTRAINT chk_client_hiring_preferences_min_lte_max
        CHECK (max_hourly_rate = 0 OR min_hourly_rate <= max_hourly_rate);
