-- +goose Up
BEGIN;

CREATE TABLE auth_identities (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL,
    provider_user_id TEXT,
    email VARCHAR(255),
    password_hash TEXT,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT auth_identities_provider_check CHECK (provider IN ('password', 'telegram')),
    CONSTRAINT auth_identities_password_check CHECK (
        provider <> 'password' OR (email IS NOT NULL AND password_hash IS NOT NULL)
    ),
    CONSTRAINT auth_identities_telegram_check CHECK (
        provider <> 'telegram' OR provider_user_id IS NOT NULL
    )
);

CREATE UNIQUE INDEX ux_auth_identities_provider_user_id
    ON auth_identities(provider, provider_user_id)
    WHERE provider_user_id IS NOT NULL;

CREATE UNIQUE INDEX ux_auth_identities_provider_email
    ON auth_identities(provider, email)
    WHERE email IS NOT NULL;

CREATE UNIQUE INDEX ux_auth_identities_primary_user
    ON auth_identities(user_id)
    WHERE is_primary;

CREATE INDEX idx_auth_identities_user_id ON auth_identities(user_id);

INSERT INTO auth_identities (user_id, provider, email, password_hash, is_primary)
SELECT id, 'password', email, password, TRUE
FROM users
WHERE email IS NOT NULL AND password IS NOT NULL;

INSERT INTO auth_identities (user_id, provider, provider_user_id, is_primary)
SELECT id, 'telegram', telegram_id::text, CASE WHEN email IS NULL THEN TRUE ELSE FALSE END
FROM users
WHERE telegram_id IS NOT NULL;

COMMIT;

-- +goose Down
BEGIN;

DROP TABLE IF EXISTS auth_identities;

COMMIT;
