-- +goose Up
BEGIN;

ALTER TABLE companies
    ADD COLUMN IF NOT EXISTS avatar_url TEXT;

ALTER TABLE events
    ADD COLUMN IF NOT EXISTS photo_url TEXT;

ALTER TABLE ideas
    ADD COLUMN IF NOT EXISTS photo_url TEXT;

COMMIT;

-- +goose Down
BEGIN;

ALTER TABLE ideas
    DROP COLUMN IF EXISTS photo_url;

ALTER TABLE events
    DROP COLUMN IF EXISTS photo_url;

ALTER TABLE companies
    DROP COLUMN IF EXISTS avatar_url;

COMMIT;
