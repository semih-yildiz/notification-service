CREATE TABLE IF NOT EXISTS notifications (
    id              TEXT PRIMARY KEY,
    batch_id        TEXT REFERENCES batches(id) ON DELETE SET NULL,
    recipient       TEXT NOT NULL,
    channel         TEXT NOT NULL CHECK (channel IN ('sms', 'email', 'push')),
    content         TEXT NOT NULL,
    priority        TEXT NOT NULL CHECK (priority IN ('high', 'normal', 'low')),
    status          TEXT NOT NULL CHECK (status IN ('pending', 'queued', 'sent', 'failed', 'cancelled')),
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at         TIMESTAMPTZ,
    failure_reason  TEXT
);

CREATE INDEX idx_notifications_batch_id ON notifications(batch_id) WHERE batch_id IS NOT NULL;
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_channel ON notifications(channel);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
CREATE UNIQUE INDEX idx_notifications_idempotency_key ON notifications(idempotency_key) WHERE idempotency_key IS NOT NULL;
