-- +goose Up
BEGIN;

-- remove references to groups first
ALTER TABLE events DROP COLUMN IF EXISTS group_id;
ALTER TABLE ideas DROP COLUMN IF EXISTS group_id;
ALTER TABLE user_availability DROP COLUMN IF EXISTS group_id;

ALTER TABLE media_archive ADD COLUMN IF NOT EXISTS company_id BIGINT REFERENCES companies(id) ON DELETE CASCADE;
ALTER TABLE media_archive DROP COLUMN IF EXISTS group_id;

ALTER TABLE events ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE ideas ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE user_availability ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE media_archive ALTER COLUMN company_id SET NOT NULL;

DROP INDEX IF EXISTS idx_groups_created_by;
DROP INDEX IF EXISTS idx_groups_invite_code;
DROP INDEX IF EXISTS idx_group_members_group;
DROP INDEX IF EXISTS idx_group_members_user;
DROP INDEX IF EXISTS idx_group_members_role;
DROP INDEX IF EXISTS idx_events_group;
DROP INDEX IF EXISTS idx_ideas_group;
DROP INDEX IF EXISTS idx_availability_user_group;
DROP INDEX IF EXISTS idx_availability_group;
DROP INDEX IF EXISTS idx_media_group;

-- drop tables last
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;

COMMIT;

-- +goose Down
ALTER TABLE media_archive ADD COLUMN IF NOT EXISTS group_id BIGINT;
ALTER TABLE media_archive DROP COLUMN IF EXISTS company_id;

ALTER TABLE user_availability ADD COLUMN IF NOT EXISTS group_id BIGINT;
ALTER TABLE ideas ADD COLUMN IF NOT EXISTS group_id BIGINT;
ALTER TABLE events ADD COLUMN IF NOT EXISTS group_id BIGINT;

CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    avatar_url TEXT,
    invite_code VARCHAR(50) UNIQUE,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS group_members (
    id SERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(group_id, user_id)
);
