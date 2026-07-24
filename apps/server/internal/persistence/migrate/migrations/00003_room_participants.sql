-- +goose Up
CREATE TABLE IF NOT EXISTS room_participants (
    id TEXT PRIMARY KEY,
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    display_name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('Player','Spectator')),
    is_host INTEGER NOT NULL DEFAULT 0 CHECK (is_host IN (0,1)),
    is_ready INTEGER NOT NULL DEFAULT 0 CHECK (is_ready IN (0,1)),
    joined_at_ms INTEGER NOT NULL,
    left_at_ms INTEGER,
    removed_at_ms INTEGER,
    removed_reason TEXT
);

CREATE INDEX IF NOT EXISTS idx_room_participants_room ON room_participants(room_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_room_participants_active_name ON room_participants(room_id, display_name) WHERE left_at_ms IS NULL AND removed_at_ms IS NULL;

-- +goose Down
DROP TABLE IF EXISTS room_participants;
