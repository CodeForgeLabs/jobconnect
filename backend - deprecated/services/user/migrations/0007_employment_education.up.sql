CREATE TABLE IF NOT EXISTS profile_employment (
    id BIGSERIAL PRIMARY KEY,
    profile_id INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    company_name VARCHAR(180) NOT NULL,
    title VARCHAR(160) NOT NULL,
    employment_type VARCHAR(40),
    location VARCHAR(180),
    is_current BOOLEAN NOT NULL DEFAULT FALSE,
    start_date DATE NOT NULL,
    end_date DATE,
    description TEXT,
    visibility VARCHAR NOT NULL DEFAULT 'PUBLIC',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT chk_profile_employment_dates
        CHECK ((is_current = TRUE AND end_date IS NULL) OR (is_current = FALSE AND end_date IS NOT NULL AND end_date >= start_date))
);

CREATE TABLE IF NOT EXISTS profile_education (
    id BIGSERIAL PRIMARY KEY,
    profile_id INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    school_name VARCHAR(180) NOT NULL,
    degree VARCHAR(160),
    field_of_study VARCHAR(160),
    start_date DATE,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT FALSE,
    grade VARCHAR(40),
    description TEXT,
    visibility VARCHAR NOT NULL DEFAULT 'PUBLIC',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT chk_profile_education_dates
        CHECK (
            (start_date IS NULL AND end_date IS NULL) OR
            (start_date IS NOT NULL AND end_date IS NULL AND is_current = TRUE) OR
            (start_date IS NOT NULL AND end_date IS NOT NULL AND end_date >= start_date)
        )
);

CREATE INDEX IF NOT EXISTS idx_profile_employment_profile_visibility_sort
    ON profile_employment(profile_id, visibility, sort_order, start_date DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_profile_education_profile_visibility_sort
    ON profile_education(profile_id, visibility, sort_order, start_date DESC)
    WHERE deleted_at IS NULL;
