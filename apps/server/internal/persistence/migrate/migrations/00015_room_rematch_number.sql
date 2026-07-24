-- +goose Up
ALTER TABLE rooms
ADD COLUMN rematch_number INTEGER NOT NULL DEFAULT 0 CHECK (rematch_number >= 0);

-- +goose Down
ALTER TABLE rooms DROP COLUMN rematch_number;
