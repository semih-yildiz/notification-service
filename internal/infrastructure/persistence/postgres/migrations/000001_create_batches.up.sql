CREATE TABLE IF NOT EXISTS batches (
    id              TEXT PRIMARY KEY,
    idempotency_key  TEXT UNIQUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_batches_idempotency_key ON batches(idempotency_key) WHERE idempotency_key IS NOT NULL;
