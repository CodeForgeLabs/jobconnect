CREATE TABLE IF NOT EXISTS freelancer_work_preferences (
    profile_id BIGINT PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    weekly_capacity_hours INTEGER NOT NULL DEFAULT 0 CHECK (weekly_capacity_hours >= 0 AND weekly_capacity_hours <= 168),
    preferred_project_length VARCHAR(50) NOT NULL DEFAULT '',
    min_budget_usd NUMERIC(10,2) NOT NULL DEFAULT 0,
    max_budget_usd NUMERIC(10,2) NOT NULL DEFAULT 0,
    contract_types JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_freelancer_work_preferences_updated_at
    ON freelancer_work_preferences(updated_at);
