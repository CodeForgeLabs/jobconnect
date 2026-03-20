ALTER TABLE freelancer_profiles
    ADD COLUMN IF NOT EXISTS job_success_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_reviews INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_jobs INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_earnings_usd DECIMAL(14,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS hourly_rate DECIMAL(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS availability VARCHAR NOT NULL DEFAULT 'AS_NEEDED',
    ADD COLUMN IF NOT EXISTS location VARCHAR,
    ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_freelancer_profiles_availability ON freelancer_profiles(availability);
CREATE INDEX IF NOT EXISTS idx_freelancer_profiles_hourly_rate ON freelancer_profiles(hourly_rate);
CREATE INDEX IF NOT EXISTS idx_freelancer_profiles_last_active_at ON freelancer_profiles(last_active_at);
