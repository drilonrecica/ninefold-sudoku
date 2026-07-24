-- +goose Up
CREATE TABLE IF NOT EXISTS room_blocks (
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    participant_id TEXT NOT NULL REFERENCES room_participants(id) ON DELETE CASCADE,
    blocked_at_ms INTEGER NOT NULL,
    reason TEXT,
    PRIMARY KEY (room_id, participant_id)
);

CREATE INDEX IF NOT EXISTS idx_room_blocks_participant ON room_blocks(participant_id);

-- +goose Down
DROP TABLE IF EXISTS room_blocks;
