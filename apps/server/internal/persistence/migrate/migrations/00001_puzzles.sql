-- +goose Up
CREATE TABLE IF NOT EXISTS puzzles (
    id TEXT NOT NULL,
    revision INTEGER NOT NULL,
    state TEXT NOT NULL CHECK (state IN ('Draft','Verified','Active','Retired')),
    difficulty TEXT NOT NULL CHECK (difficulty IN ('Easy','Medium','Hard','Expert','Random')),
    hardest_technique TEXT NOT NULL,
    quality_score REAL NOT NULL,
    multiplayer_approved INTEGER NOT NULL CHECK (multiplayer_approved IN (0,1)),
    generator_version TEXT NOT NULL,
    solver_version TEXT NOT NULL,
    canonical_fingerprint TEXT NOT NULL,
    clues BLOB NOT NULL CHECK (length(clues) = 81),
    solution BLOB NOT NULL CHECK (length(solution) = 81),
    created_at_ms INTEGER NOT NULL,
    PRIMARY KEY (id, revision)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_puzzles_fingerprint ON puzzles(canonical_fingerprint);
CREATE INDEX IF NOT EXISTS idx_puzzles_active_difficulty ON puzzles(state, difficulty);

-- +goose Down
DROP TABLE IF EXISTS puzzles;
