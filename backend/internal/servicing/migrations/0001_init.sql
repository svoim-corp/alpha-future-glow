-- +goose Up
-- Выданные кредиты и графики платежей (docs/servicing.md). Кредит заводится при выдаче, один на заявку.
CREATE SEQUENCE servicing.loan_seq;

CREATE TABLE servicing.loans (
    id              text PRIMARY KEY,
    app_id          text          NOT NULL UNIQUE,
    company_id      text          NOT NULL,
    applicant_id    text          NOT NULL,
    contract_number text          NOT NULL,
    product_code    text          NOT NULL,
    product_name    text          NOT NULL,
    amount          bigint        NOT NULL CHECK (amount > 0),
    term            int           NOT NULL CHECK (term > 0),
    rate            numeric(5, 2) NOT NULL,
    grace_months    int           NOT NULL DEFAULT 0,
    monthly_payment bigint        NOT NULL,
    issued_at       timestamptz   NOT NULL,
    status          text          NOT NULL CHECK (status IN ('ACTIVE', 'CLOSED'))
);

CREATE INDEX loans_company ON servicing.loans (company_id, issued_at DESC, id);

CREATE TABLE servicing.payments (
    loan_id   text   NOT NULL REFERENCES servicing.loans (id) ON DELETE CASCADE,
    n         int    NOT NULL,
    due_date  date   NOT NULL,
    principal bigint NOT NULL,
    interest  bigint NOT NULL,
    total     bigint NOT NULL,
    status    text   NOT NULL CHECK (status IN ('PLANNED', 'PAID')),
    PRIMARY KEY (loan_id, n)
);

CREATE INDEX payments_due ON servicing.payments (due_date, loan_id);

-- +goose Down
DROP TABLE servicing.payments;
DROP TABLE servicing.loans;
DROP SEQUENCE servicing.loan_seq;
