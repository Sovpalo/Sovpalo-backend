-- +goose Up
BEGIN;

CREATE TABLE idea_likes (
    id SERIAL PRIMARY KEY,
    idea_id BIGINT NOT NULL REFERENCES ideas(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(idea_id, user_id)
);

CREATE INDEX idx_idea_likes_idea ON idea_likes(idea_id);
CREATE INDEX idx_idea_likes_user ON idea_likes(user_id);

COMMIT;

-- +goose Down
DROP TABLE IF EXISTS idea_likes;
