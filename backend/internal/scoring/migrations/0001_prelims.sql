-- +goose Up
-- Результаты этапа 1 (Prelim, API_SPEC.md, раздел 5). Повторная проверка добавляет строку, действует последняя.
-- product_code, amount, term — параметры, по которым посчитан результат: Origination сравнивает их с заявкой.
CREATE TABLE scoring.prelims (
    id             bigserial PRIMARY KEY,
    app_id         text NOT NULL,
    outcome        text NOT NULL CHECK (outcome IN ('LIMIT', 'STOP')),
    limit_amount   bigint NOT NULL DEFAULT 0,
    prelim_rate    double precision NOT NULL DEFAULT 0,
    prelim_payment bigint NOT NULL DEFAULT 0,
    amount_fits    boolean NOT NULL DEFAULT false,
    stop_reason    text CHECK (stop_reason IN ('COMPANY_INACTIVE', 'BUSINESS_TOO_YOUNG', 'BKI_OVERDUE', 'CRITICAL_RISK')),
    choice         text CHECK (choice IN ('REDUCE_AMOUNT', 'KEEP_AMOUNT')),
    product_code   text NOT NULL,
    amount         bigint NOT NULL,
    term           int NOT NULL,
    checks         jsonb NOT NULL DEFAULT '[]',
    created_at     timestamptz NOT NULL,
    CHECK ((outcome = 'STOP') = (stop_reason IS NOT NULL))
);

CREATE INDEX prelims_latest ON scoring.prelims (app_id, created_at DESC, id DESC);

-- +goose Down
DROP TABLE scoring.prelims;
