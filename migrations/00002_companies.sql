-- +goose Up
BEGIN;

CREATE TABLE companies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE company_members (
    id SERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, user_id)
);

CREATE TABLE company_invitations (
    id SERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    invited_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invited_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    responded_at TIMESTAMPTZ,
    UNIQUE(company_id, invited_user_id),
    CONSTRAINT company_invitations_status_check CHECK (status IN ('pending', 'accepted', 'declined'))
);

CREATE INDEX idx_companies_created_by ON companies(created_by);
CREATE INDEX idx_company_members_company ON company_members(company_id);
CREATE INDEX idx_company_members_user ON company_members(user_id);
CREATE INDEX idx_company_invitations_user ON company_invitations(invited_user_id);
CREATE INDEX idx_company_invitations_status ON company_invitations(status);
CREATE INDEX idx_company_invitations_company ON company_invitations(company_id);

COMMIT;

-- +goose Down
DROP TABLE IF EXISTS company_invitations;
DROP TABLE IF EXISTS company_members;
DROP TABLE IF EXISTS companies;
