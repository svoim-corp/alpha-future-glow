-- +goose Up
-- Очередь банка — своя копия нужных полей заявки, обновляется по ApplicationStatusChanged:
-- JOIN со схемами других зон запрещён, а фильтры и сортировка должны работать в одном запросе.
CREATE TABLE underwriting.queue (
    app_id        text PRIMARY KEY,
    number        text,
    company_id    text NOT NULL,
    company_inn   text NOT NULL,
    company_name  text NOT NULL,
    product_code  text NOT NULL,
    product_name  text NOT NULL,
    amount        bigint NOT NULL,
    term          int NOT NULL,
    status        text NOT NULL,
    score         int,
    risk_band     text CHECK (risk_band IN ('LOW', 'MEDIUM', 'HIGH')),
    decision_mode text CHECK (decision_mode IN ('AUTO', 'MANUAL')),
    assignee_id   text,
    assignee_name text,
    submitted_at  timestamptz NOT NULL,
    decided_at    timestamptz,
    updated_at    timestamptz NOT NULL
);

CREATE INDEX queue_status ON underwriting.queue (status, submitted_at DESC);
CREATE INDEX queue_decided ON underwriting.queue (decided_at) WHERE decided_at IS NOT NULL;

-- +goose Down
DROP TABLE underwriting.queue;
