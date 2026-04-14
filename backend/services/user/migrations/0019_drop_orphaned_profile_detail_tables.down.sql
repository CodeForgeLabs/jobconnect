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

CREATE TABLE IF NOT EXISTS profile_certifications (
    id BIGSERIAL PRIMARY KEY,
    profile_id INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name VARCHAR(180) NOT NULL,
    issuing_organization VARCHAR(180) NOT NULL,
    credential_id VARCHAR(160),
    credential_url VARCHAR(2048),
    issue_date DATE,
    expiration_date DATE,
    does_not_expire BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT chk_profile_cert_dates
        CHECK (
            (does_not_expire = TRUE AND expiration_date IS NULL) OR
            (does_not_expire = FALSE AND (expiration_date IS NULL OR issue_date IS NULL OR expiration_date >= issue_date))
        )
);

CREATE TABLE IF NOT EXISTS profile_languages (
    profile_id INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    proficiency VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (profile_id, language_code),
    CONSTRAINT chk_profile_languages_proficiency
        CHECK (proficiency IN ('BASIC', 'CONVERSATIONAL', 'FLUENT', 'NATIVE'))
);
