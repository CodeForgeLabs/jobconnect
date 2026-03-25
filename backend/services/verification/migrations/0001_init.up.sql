CREATE TABLE IF NOT EXISTS verification_requests (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    request_version INT NOT NULL,
    status VARCHAR NOT NULL,
    legal_name VARCHAR NOT NULL,
    country_code VARCHAR(2) NOT NULL,
    document_type VARCHAR NOT NULL,
    document_number_masked VARCHAR NOT NULL,
    evidence_urls JSONB NOT NULL DEFAULT '[]'::jsonb,
    submission_note TEXT,
    reviewer_user_id UUID,
    rejection_reason TEXT,
    internal_note TEXT,
    submitted_at TIMESTAMP NOT NULL,
    reviewed_at TIMESTAMP,
    reverify_due_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE (user_id, request_version)
);

CREATE INDEX IF NOT EXISTS idx_verification_requests_status_submitted_at
    ON verification_requests(status, submitted_at);

CREATE INDEX IF NOT EXISTS idx_verification_requests_user_id
    ON verification_requests(user_id);

CREATE INDEX IF NOT EXISTS idx_verification_requests_reverify_due_at
    ON verification_requests(reverify_due_at);

CREATE TABLE IF NOT EXISTS verification_events (
    id BIGSERIAL PRIMARY KEY,
    request_id BIGINT NOT NULL REFERENCES verification_requests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    event_type VARCHAR NOT NULL,
    actor_user_id UUID,
    details_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_verification_events_request_id_created_at
    ON verification_events(request_id, created_at);
