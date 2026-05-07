-- +goose Up
BEGIN;

CREATE TABLE push_device_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    platform VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT push_device_tokens_platform_check CHECK (platform IN ('ios'))
);

CREATE INDEX idx_push_device_tokens_user_id ON push_device_tokens(user_id);

COMMIT;

-- +goose Down
BEGIN;

DROP TABLE IF EXISTS push_device_tokens;

COMMIT;
