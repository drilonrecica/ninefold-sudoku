-- +goose Up
CREATE TABLE IF NOT EXISTS match_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    match_id TEXT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    event_number INTEGER NOT NULL,
    aggregate_version INTEGER NOT NULL,
    state_format TEXT NOT NULL,
    state_blob BLOB NOT NULL,
    integrity_hash BLOB NOT NULL,
    created_at_ms INTEGER NOT NULL,
    UNIQUE (match_id, event_number)
);

CREATE INDEX IF NOT EXISTS idx_match_snapshots_match ON match_snapshots(match_id);

-- +goose Down
DROP TABLE IF EXISTS match_snapshots;
