-- +goose Up
BEGIN;

ALTER TABLE user_availability
    ADD COLUMN company_id BIGINT REFERENCES companies(id) ON DELETE CASCADE;

ALTER TABLE user_availability
    ALTER COLUMN group_id DROP NOT NULL;

ALTER TABLE user_availability
    ADD CONSTRAINT user_availability_scope_check CHECK (
        group_id IS NOT NULL OR company_id IS NOT NULL
    );

CREATE INDEX idx_availability_company ON user_availability(company_id);

COMMIT;

-- +goose Down
DROP INDEX IF EXISTS idx_availability_company;
ALTER TABLE user_availability
    DROP CONSTRAINT IF EXISTS user_availability_scope_check;
ALTER TABLE user_availability
    ALTER COLUMN group_id SET NOT NULL;
ALTER TABLE user_availability
    DROP COLUMN IF EXISTS company_id;
