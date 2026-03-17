CREATE TABLE IF NOT EXISTS reviews (
    id              BIGSERIAL PRIMARY KEY,
    contract_id     BIGINT NOT NULL,
    reviewer_id     UUID NOT NULL,
    reviewee_id     UUID NOT NULL,
    reviewer_role   TEXT NOT NULL CHECK (reviewer_role IN ('client', 'freelancer')),
    rating          INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title           TEXT NOT NULL,
    comment         TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL,

    CONSTRAINT unique_review_per_contract UNIQUE (contract_id, reviewer_id)
);

CREATE INDEX IF NOT EXISTS idx_reviews_reviewee ON reviews(reviewee_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reviews_contract ON reviews(contract_id);
