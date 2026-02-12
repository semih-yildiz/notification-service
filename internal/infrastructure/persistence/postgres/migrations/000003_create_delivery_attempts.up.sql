CREATE TABLE IF NOT EXISTS delivery_attempts (
    id              TEXT PRIMARY KEY,
    notification_id TEXT NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    attempt_number  INT NOT NULL,
    success         BOOLEAN NOT NULL,
    status_code     INT NOT NULL,
    response_body   TEXT,
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_delivery_attempts_notification_id ON delivery_attempts(notification_id);
