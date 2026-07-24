-- +goose Up
CREATE TABLE IF NOT EXISTS replay_capabilities (
    token_hash BLOB PRIMARY KEY,
    replay_id TEXT NOT NULL,
    match_id TEXT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    created_at_ms INTEGER NOT NULL,
    expires_at_ms INTEGER NOT NULL,
    revoked_at_ms INTEGER
);

CREATE INDEX IF NOT EXISTS idx_replay_capabilities_match ON replay_capabilities(match_id);
CREATE INDEX IF NOT EXISTS idx_replay_capabilities_expires ON replay_capabilities(expires_at_ms);

CREATE TABLE IF NOT EXISTS replay_seals (
    match_id TEXT PRIMARY KEY REFERENCES matches(id) ON DELETE CASCADE,
    final_event_number INTEGER NOT NULL,
    final_event_hash BLOB NOT NULL,
    terminal_at_ms INTEGER NOT NULL,
    signing_key_id TEXT NOT NULL,
    signature BLOB NOT NULL,
    proof_version TEXT NOT NULL,
    created_at_ms INTEGER NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS replay_capabilities;
DROP TABLE IF EXISTS replay_seals;
