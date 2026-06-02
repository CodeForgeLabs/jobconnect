DROP INDEX IF EXISTS idx_freelancer_profiles_last_active_at;
DROP INDEX IF EXISTS idx_freelancer_profiles_hourly_rate;
DROP INDEX IF EXISTS idx_freelancer_profiles_availability;

ALTER TABLE freelancer_profiles
    DROP COLUMN IF EXISTS last_active_at,
    DROP COLUMN IF EXISTS location,
    DROP COLUMN IF EXISTS availability,
    DROP COLUMN IF EXISTS hourly_rate,
    DROP COLUMN IF EXISTS total_earnings_usd,
    DROP COLUMN IF EXISTS total_jobs,
    DROP COLUMN IF EXISTS total_reviews,
    DROP COLUMN IF EXISTS job_success_score;
