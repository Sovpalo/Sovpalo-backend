-- +goose Up
BEGIN;

ALTER TABLE events
    ADD COLUMN company_id BIGINT REFERENCES companies(id) ON DELETE SET NULL;

ALTER TABLE events
    ALTER COLUMN group_id DROP NOT NULL;

CREATE INDEX idx_events_company ON events(company_id);

COMMIT;

-- +goose Down
DROP INDEX IF EXISTS idx_events_company;
ALTER TABLE events
    ALTER COLUMN group_id SET NOT NULL;
ALTER TABLE events
    DROP COLUMN IF EXISTS company_id;
