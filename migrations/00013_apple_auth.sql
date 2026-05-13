-- +goose Up
BEGIN;

ALTER TABLE auth_identities DROP CONSTRAINT auth_identities_provider_check;
ALTER TABLE auth_identities ADD CONSTRAINT auth_identities_provider_check
    CHECK (provider IN ('password', 'telegram', 'apple'));

ALTER TABLE auth_identities ADD CONSTRAINT auth_identities_apple_check CHECK (
    provider <> 'apple' OR provider_user_id IS NOT NULL
);

COMMIT;

-- +goose Down
BEGIN;

ALTER TABLE auth_identities DROP CONSTRAINT auth_identities_apple_check;

ALTER TABLE auth_identities DROP CONSTRAINT auth_identities_provider_check;
ALTER TABLE auth_identities ADD CONSTRAINT auth_identities_provider_check
    CHECK (provider IN ('password', 'telegram'));

COMMIT;
