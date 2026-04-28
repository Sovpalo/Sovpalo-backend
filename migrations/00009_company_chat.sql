-- +goose Up
BEGIN;

CREATE TABLE company_chat_messages (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    sender_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT company_chat_messages_text_not_blank CHECK (text IS NULL OR btrim(text) <> '')
);

CREATE TABLE company_chat_attachments (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL REFERENCES company_chat_messages(id) ON DELETE CASCADE,
    file_name VARCHAR(500) NOT NULL,
    file_url TEXT NOT NULL,
    file_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    media_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT company_chat_attachments_media_type_check CHECK (media_type IN ('photo', 'video'))
);

CREATE TABLE company_chat_message_reads (
    message_id BIGINT NOT NULL REFERENCES company_chat_messages(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (message_id, user_id)
);

CREATE INDEX idx_company_chat_messages_company_id_id_desc
    ON company_chat_messages(company_id, id DESC);
CREATE INDEX idx_company_chat_messages_sender_id_created_at
    ON company_chat_messages(sender_id, created_at DESC);
CREATE INDEX idx_company_chat_attachments_message_id
    ON company_chat_attachments(message_id);
CREATE INDEX idx_company_chat_message_reads_user_id_message_id
    ON company_chat_message_reads(user_id, message_id);

COMMIT;

-- +goose Down
BEGIN;

DROP TABLE IF EXISTS company_chat_message_reads;
DROP TABLE IF EXISTS company_chat_attachments;
DROP TABLE IF EXISTS company_chat_messages;

COMMIT;
