-- +goose Up
-- Запросы информации у клиента (InfoRequest). Банк спрашивает только пояснения: данные и документы у него уже есть (Уточн.).
CREATE SEQUENCE underwriting.info_request_seq;

CREATE TABLE underwriting.info_requests (
    id          text PRIMARY KEY,
    app_id      text NOT NULL,
    comment     text NOT NULL,
    items       jsonb NOT NULL,
    answers     jsonb,
    status      text NOT NULL CHECK (status IN ('OPEN', 'ANSWERED')),
    created_by  text NOT NULL,
    created_at  timestamptz NOT NULL,
    answered_at timestamptz,
    CHECK ((status = 'ANSWERED') = (answers IS NOT NULL))
);

-- Открытый запрос по заявке — не больше одного: клиент отвечает на текущий, а не выбирает из нескольких.
CREATE UNIQUE INDEX info_requests_one_open ON underwriting.info_requests (app_id) WHERE status = 'OPEN';
CREATE INDEX info_requests_app ON underwriting.info_requests (app_id, created_at DESC, id);

-- +goose Down
DROP TABLE underwriting.info_requests;
DROP SEQUENCE underwriting.info_request_seq;
