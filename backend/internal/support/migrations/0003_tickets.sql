-- +goose Up
-- Обращения клиента. Со страницы заявки обращение привязывается к ней.
CREATE SEQUENCE support.ticket_seq;

CREATE TABLE support.tickets (
    id             text PRIMARY KEY,
    user_id        text        NOT NULL,
    company_id     text        NOT NULL,
    application_id text,
    subject        text        NOT NULL,
    text           text        NOT NULL,
    status         text        NOT NULL CHECK (status IN ('OPEN', 'ANSWERED', 'CLOSED')),
    reply          text,
    created_at     timestamptz NOT NULL,
    updated_at     timestamptz NOT NULL
);

CREATE INDEX tickets_user ON support.tickets (user_id, updated_at DESC, id);

-- +goose Down
DROP TABLE support.tickets;
DROP SEQUENCE support.ticket_seq;
