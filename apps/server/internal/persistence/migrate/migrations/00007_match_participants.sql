-- +goose Up
CREATE TABLE IF NOT EXISTS match_participants (
    match_id TEXT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    participant_id TEXT NOT NULL REFERENCES room_participants(id) ON DELETE CASCADE,
    connected INTEGER NOT NULL DEFAULT 1 CHECK (connected IN (0,1)),
    mistakes INTEGER NOT NULL DEFAULT 0,
    hints_used INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (match_id, participant_id)
);

CREATE INDEX IF NOT EXISTS idx_match_participants_match ON match_participants(match_id);

-- +goose Down
DROP TABLE IF EXISTS match_participants;
