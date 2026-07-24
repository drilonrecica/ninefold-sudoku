-- +goose Up
CREATE TABLE IF NOT EXISTS rooms (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    state TEXT NOT NULL CHECK (state IN ('Lobby','Countdown','InMatch','Results','Expired','Cancelled','RecoveryPending','TerminatedByAdmin')),
    version INTEGER NOT NULL,
    mode TEXT NOT NULL CHECK (mode IN ('Coop','Race','Duel')),
    difficulty TEXT NOT NULL,
    error_preset TEXT NOT NULL CHECK (error_preset IN ('Casual','Challenge','Blind','Clean')),
    hints_enabled INTEGER NOT NULL CHECK (hints_enabled IN (0,1)),
    shared_notes INTEGER NOT NULL CHECK (shared_notes IN (0,1)),
    auto_remove_notes INTEGER NOT NULL CHECK (auto_remove_notes IN (0,1)),
    spectators_allowed INTEGER NOT NULL CHECK (spectators_allowed IN (0,1)),
    host_participant_id TEXT NOT NULL,
    current_match_id TEXT,
    created_at_ms INTEGER NOT NULL,
    last_activity_at_ms INTEGER NOT NULL,
    expires_at_ms INTEGER NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_rooms_code ON rooms(code);

-- +goose Down
DROP TABLE IF EXISTS rooms;
