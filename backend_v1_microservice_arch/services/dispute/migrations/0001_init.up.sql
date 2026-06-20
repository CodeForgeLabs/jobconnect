CREATE TABLE IF NOT EXISTS disputes (
  id BIGSERIAL PRIMARY KEY,
  reference_type VARCHAR(64) NOT NULL,
  reference_id VARCHAR(255) NOT NULL,
  opened_by UUID NOT NULL,
  reason TEXT NOT NULL,
  status VARCHAR(32) NOT NULL,
  decision VARCHAR(32) NOT NULL DEFAULT '',
  resolution_note TEXT NOT NULL DEFAULT '',
  resolved_by UUID NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  resolved_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_disputes_reference ON disputes(reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_disputes_status ON disputes(status);

