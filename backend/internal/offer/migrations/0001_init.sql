-- +goose Up
-- Предложение банка по решению и договор (docs/offer.md). Одна строка на заявку: новое решение по заявке заменяет
-- предложение, пока клиент его не принял.
CREATE TABLE offer.offers (
    app_id                 text PRIMARY KEY,
    decision_id            text          NOT NULL,
    decision               text          NOT NULL CHECK (decision IN ('APPROVED', 'COUNTER_OFFER', 'REJECTED')),
    product_code           text          NOT NULL,
    amount                 bigint        NOT NULL,
    term                   int           NOT NULL,
    rate                   numeric(5, 2) NOT NULL,
    grace_months           int,
    monthly_payment        bigint        NOT NULL,
    requested_product_code text          NOT NULL,
    requested_amount       bigint        NOT NULL,
    requested_term         int           NOT NULL,
    explanation            text,
    alternatives           jsonb         NOT NULL,
    valid_until            date          NOT NULL,
    created_at             timestamptz   NOT NULL
);

-- +goose Down
DROP TABLE offer.offers;
