-- +goose Up
CREATE TABLE IF NOT EXISTS matches (
    id TEXT PRIMARY KEY,
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    state TEXT NOT NULL CHECK (state IN ('Prepared','Countdown','Active','Finishing','Completed','RecoveryPending','Cancelled','Abandoned','Invalidated')),
    version INTEGER NOT NULL,
    mode TEXT NOT NULL CHECK (mode IN ('Coop','Race','Duel')),
    difficulty TEXT NOT NULL,
    error_preset TEXT NOT NULL CHECK (error_preset IN ('Casual','Challenge','Blind','Clean')),
    hints_enabled INTEGER NOT NULL CHECK (hints_enabled IN (0,1)),
    auto_remove_notes INTEGER NOT NULL CHECK (auto_remove_notes IN (0,1)),
    rule_version INTEGER NOT NULL,
    puzzle_id TEXT NOT NULL,
    puzzle_revision INTEGER NOT NULL,
    transformation_seed INTEGER NOT NULL,
    puzzle_difficulty TEXT NOT NULL,
    generator_version TEXT NOT NULL,
    solver_version TEXT NOT NULL,
    started_at_ms INTEGER,
    completed_at_ms INTEGER,
    result_reason TEXT,
    elapsed_ms INTEGER,
    assisted INTEGER NOT NULL DEFAULT 0 CHECK (assisted IN (0,1)),
    created_at_ms INTEGER NOT NULL,
    FOREIGN KEY (puzzle_id, puzzle_revision) REFERENCES puzzles(id, revision) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_matches_room ON matches(room_id);
CREATE INDEX IF NOT EXISTS idx_matches_room_state ON matches(room_id, state);

-- +goose Down
DROP TABLE IF EXISTS matches;
