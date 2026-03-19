-- +goose Up
BEGIN;

ALTER TABLE ideas
    ADD COLUMN company_id BIGINT REFERENCES companies(id) ON DELETE CASCADE;

ALTER TABLE ideas
    ALTER COLUMN group_id DROP NOT NULL;

ALTER TABLE ideas
    ADD CONSTRAINT ideas_scope_check CHECK (
        group_id IS NOT NULL OR company_id IS NOT NULL
    );

CREATE INDEX idx_ideas_company ON ideas(company_id);

COMMIT;

-- +goose Down
DROP INDEX IF EXISTS idx_ideas_company;
ALTER TABLE ideas
    DROP CONSTRAINT IF EXISTS ideas_scope_check;
ALTER TABLE ideas
    ALTER COLUMN group_id SET NOT NULL;
ALTER TABLE ideas
    DROP COLUMN IF EXISTS company_id;
