-- +goose Up
-- Задачи конвейера после отправки заявки. Хранятся в БД, чтобы рестарт сервера не терял заявки в работе.
CREATE TABLE scoring.jobs (
    app_id     text PRIMARY KEY,
    state      text NOT NULL CHECK (state IN ('PENDING', 'DONE', 'FAILED')),
    attempts   int NOT NULL DEFAULT 0,
    last_error text,
    updated_at timestamptz NOT NULL
);

CREATE INDEX jobs_pending ON scoring.jobs (updated_at) WHERE state = 'PENDING';

-- Результаты полного скоринга (ScoreResult). Пересчёт добавляет строку, действует последняя.
CREATE TABLE scoring.score_results (
    id            bigserial PRIMARY KEY,
    app_id        text NOT NULL,
    score         int NOT NULL,
    risk_band     text NOT NULL CHECK (risk_band IN ('LOW', 'MEDIUM', 'HIGH')),
    factors       jsonb NOT NULL,
    auto_eligible boolean NOT NULL,
    auto_reason   text NOT NULL,
    suggested     jsonb NOT NULL,
    model_version text NOT NULL,
    created_at    timestamptz NOT NULL
);

CREATE INDEX score_results_latest ON scoring.score_results (app_id, created_at DESC, id DESC);

-- Решения по заявке: авто (decided_by IS NULL) и сотрудника. Действует последнее.
CREATE SEQUENCE scoring.decision_seq;

CREATE TABLE scoring.decisions (
    id           text PRIMARY KEY,
    app_id       text NOT NULL,
    result       text NOT NULL CHECK (result IN ('APPROVED', 'COUNTER_OFFER', 'REJECTED')),
    mode         text NOT NULL CHECK (mode IN ('AUTO', 'MANUAL')),
    product_code text NOT NULL,
    amount       bigint NOT NULL,
    term         int NOT NULL,
    rate         double precision NOT NULL,
    reason       text NOT NULL,
    alternatives jsonb NOT NULL DEFAULT '[]',
    decided_by   text,
    decided_at   timestamptz NOT NULL,
    CHECK ((mode = 'AUTO') = (decided_by IS NULL))
);

CREATE INDEX decisions_latest ON scoring.decisions (app_id, decided_at DESC, id DESC);

-- +goose Down
DROP TABLE scoring.decisions;
DROP SEQUENCE scoring.decision_seq;
DROP TABLE scoring.score_results;
DROP TABLE scoring.jobs;
