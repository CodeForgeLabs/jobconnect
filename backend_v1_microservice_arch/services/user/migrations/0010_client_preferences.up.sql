CREATE TABLE IF NOT EXISTS client_hiring_preferences (
    profile_id BIGINT PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    min_hourly_rate NUMERIC(10,2) NOT NULL DEFAULT 0,
    max_hourly_rate NUMERIC(10,2) NOT NULL DEFAULT 0,
    preferred_experience_levels JSONB NOT NULL DEFAULT '[]'::jsonb,
    preferred_locations JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS client_saved_freelancers (
    profile_id BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    freelancer_user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (profile_id, freelancer_user_id)
);

CREATE TABLE IF NOT EXISTS client_freelancer_notes (
    profile_id BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    freelancer_user_id UUID NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (profile_id, freelancer_user_id)
);

CREATE INDEX IF NOT EXISTS idx_client_saved_freelancers_created_at
    ON client_saved_freelancers(profile_id, created_at DESC);
