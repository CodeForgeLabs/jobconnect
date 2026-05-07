ALTER TABLE verification_requests
    ADD COLUMN IF NOT EXISTS evidence_url TEXT;

UPDATE verification_requests
SET evidence_url = COALESCE(NULLIF(TRIM(evidence_urls->>0), ''), '')
WHERE evidence_url IS NULL;

UPDATE verification_requests
SET evidence_url = ''
WHERE evidence_url IS NULL;

ALTER TABLE verification_requests
    ALTER COLUMN evidence_url SET DEFAULT '';

ALTER TABLE verification_requests
    ALTER COLUMN evidence_url SET NOT NULL;

ALTER TABLE verification_requests
    DROP COLUMN IF EXISTS evidence_urls;

