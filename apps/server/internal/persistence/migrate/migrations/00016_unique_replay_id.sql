-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS idx_replay_capabilities_replay_id
    ON replay_capabilities(replay_id);

-- +goose Down
DROP INDEX IF EXISTS idx_replay_capabilities_replay_id;
