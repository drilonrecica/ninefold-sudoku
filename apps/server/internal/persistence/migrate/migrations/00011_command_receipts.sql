-- +goose Up
CREATE TABLE IF NOT EXISTS command_receipts (
    request_id TEXT PRIMARY KEY,
    authenticated_scope_hash BLOB NOT NULL,
    command_type TEXT NOT NULL,
    request_fingerprint TEXT NOT NULL,
    terminal_status TEXT NOT NULL,
    safe_response_json TEXT,
    created_at_ms INTEGER NOT NULL,
    expires_at_ms INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_command_receipts_expires ON command_receipts(expires_at_ms);

-- +goose Down
DROP TABLE IF EXISTS command_receipts;
