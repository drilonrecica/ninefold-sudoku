-- +goose Up
CREATE TABLE IF NOT EXISTS room_sessions (
    token_hash BLOB PRIMARY KEY,
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    participant_id TEXT NOT NULL REFERENCES room_participants(id) ON DELETE CASCADE,
    created_at_ms INTEGER NOT NULL,
    expires_at_ms INTEGER NOT NULL,
    revoked_at_ms INTEGER
);

CREATE INDEX IF NOT EXISTS idx_room_sessions_participant ON room_sessions(participant_id);
CREATE INDEX IF NOT EXISTS idx_room_sessions_expires ON room_sessions(expires_at_ms);

-- +goose Down
DROP TABLE IF EXISTS room_sessions;
