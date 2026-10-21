-- +goose Up
-- Снимки ответов внешних источников по заявке (CheckResult, API_SPEC.md, раздел 5). Запуск проверок только добавляет строки:
-- история перезапросов остаётся, актуальный снимок — последний по источнику.
CREATE TABLE integrations.check_results (
    id           bigserial PRIMARY KEY,
    app_id       text NOT NULL,
    source       text NOT NULL CHECK (source IN ('BANK', 'BANK_ACCOUNTS', 'EGRUL', 'TRANSPARENT_BUSINESS', 'BKI', 'FNS', 'GIRBO', 'COURTS')),
    status       text NOT NULL CHECK (status IN ('OK', 'WARN', 'FAIL', 'UNAVAILABLE')),
    requested_at timestamptz NOT NULL,
    summary      text NOT NULL,
    details      jsonb NOT NULL DEFAULT '{}',
    payload      jsonb
);

CREATE INDEX check_results_latest ON integrations.check_results (app_id, source, requested_at DESC, id DESC);

-- +goose Down
DROP TABLE integrations.check_results;
