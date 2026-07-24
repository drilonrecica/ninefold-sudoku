-- +goose Up
CREATE TABLE IF NOT EXISTS match_events (
    match_id TEXT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    event_number INTEGER NOT NULL,
    aggregate_version INTEGER NOT NULL,
    public_event_type TEXT NOT NULL,
    public_actor_id TEXT,
    request_id TEXT NOT NULL,
    occurred_at_ms INTEGER NOT NULL,
    public_payload_json TEXT NOT NULL,
    private_payload_blob BLOB,
    private_payload_salt BLOB,
    private_payload_digest BLOB,
    previous_hash BLOB,
    event_hash BLOB,
    PRIMARY KEY (match_id, event_number),
    UNIQUE (match_id, request_id)
);

CREATE INDEX IF NOT EXISTS idx_match_events_match ON match_events(match_id);

-- +goose Down
DROP TABLE IF EXISTS match_events;
