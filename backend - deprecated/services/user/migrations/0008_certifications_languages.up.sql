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
    visibility VARCHAR NOT NULL DEFAULT 'PUBLIC',
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
    visibility VARCHAR NOT NULL DEFAULT 'PUBLIC',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (profile_id, language_code),
    CONSTRAINT chk_profile_languages_proficiency
        CHECK (proficiency IN ('BASIC', 'CONVERSATIONAL', 'FLUENT', 'NATIVE'))
);

CREATE INDEX IF NOT EXISTS idx_profile_certifications_profile_visibility
    ON profile_certifications(profile_id, visibility, issue_date DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_profile_languages_profile_visibility
    ON profile_languages(profile_id, visibility);
