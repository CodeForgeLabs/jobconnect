CREATE TABLE connects_balances (
    user_id UUID PRIMARY KEY,
    balance INT NOT NULL CHECK (balance >= 0),
    version INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE connects_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    amount INT NOT NULL, -- positive for credit, negative for debit
    type TEXT NOT NULL, -- e.g. PROPOSAL_SUBMITTED, REGISTER_BONUS, FIAT_PURCHASE
    reference_id TEXT NOT NULL,
    reference_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for querying transaction ledger for a user
CREATE INDEX idx_connects_transactions_user_id ON connects_transactions(user_id, created_at DESC);

-- Unique index to prevent idempotency issues (same reference_type and reference_id for a user should only have ONE transaction)
-- For example: reference_type='PROPOSAL_SUBMITTED', reference_id='proposal_uuid'
-- This ensures if we retry the deduction, we don't insert a second row.
CREATE UNIQUE INDEX idx_connects_transactions_idempotency ON connects_transactions(user_id, reference_type, reference_id);
