ALTER TABLE verification_requests
    ADD COLUMN IF NOT EXISTS evidence_urls JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE verification_requests
SET evidence_urls = CASE
    WHEN COALESCE(TRIM(evidence_url), '') = '' THEN '[]'::jsonb
    ELSE jsonb_build_array(evidence_url)
END;

ALTER TABLE verification_requests
    DROP COLUMN IF EXISTS evidence_url;

