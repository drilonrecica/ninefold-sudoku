-- +goose Up
CREATE TABLE IF NOT EXISTS match_results (
    match_id TEXT PRIMARY KEY REFERENCES matches(id) ON DELETE CASCADE,
    result_reason TEXT NOT NULL,
    elapsed_ms INTEGER NOT NULL,
    assisted INTEGER NOT NULL DEFAULT 0 CHECK (assisted IN (0,1)),
    created_at_ms INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS match_result_players (
    match_id TEXT NOT NULL REFERENCES match_results(match_id) ON DELETE CASCADE,
    participant_id TEXT NOT NULL,
    display_name TEXT NOT NULL,
    mistakes INTEGER NOT NULL DEFAULT 0,
    hints_used INTEGER NOT NULL DEFAULT 0,
    score INTEGER,
    PRIMARY KEY (match_id, participant_id)
);

CREATE INDEX IF NOT EXISTS idx_match_result_players_match ON match_result_players(match_id);

CREATE TABLE IF NOT EXISTS match_tombstones (
    match_id TEXT PRIMARY KEY,
    mode TEXT NOT NULL,
    difficulty TEXT NOT NULL,
    result_reason TEXT NOT NULL,
    started_at_ms INTEGER,
    ended_at_ms INTEGER NOT NULL,
    schema_version TEXT NOT NULL,
    proof_version TEXT NOT NULL,
    replay_deleted_at_ms INTEGER,
    replay_expired_at_ms INTEGER,
    created_at_ms INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_match_tombstones_expires ON match_tombstones(ended_at_ms);

-- +goose Down
DROP TABLE IF EXISTS match_result_players;
DROP TABLE IF EXISTS match_results;
DROP TABLE IF EXISTS match_tombstones;
