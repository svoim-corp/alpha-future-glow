-- +goose Up
-- Оплаты, досрочное погашение и закрытие (docs/servicing.md). Плановые суммы строк не меняются: фактически
-- оплаченное хранится рядом, проводки — в журнале, заменённые перерасчётом строки — в архиве версий графика.
ALTER TABLE servicing.loans
    ADD COLUMN version          int         NOT NULL DEFAULT 1,
    ADD COLUMN schedule_version int         NOT NULL DEFAULT 1,
    ADD COLUMN revolving        boolean     NOT NULL DEFAULT false,
    ADD COLUMN penalties_paid   bigint      NOT NULL DEFAULT 0 CHECK (penalties_paid >= 0),
    ADD COLUMN closed_at        timestamptz,
    ADD COLUMN close_reason     text CHECK (close_reason IN ('SCHEDULE_COMPLETED', 'EARLY_REPAYMENT', 'LIMIT_CLOSED')),
    ADD CONSTRAINT loans_closed CHECK ((status = 'CLOSED') = (closed_at IS NOT NULL AND close_reason IS NOT NULL));

ALTER TABLE servicing.payments
    ADD COLUMN paid_principal bigint  NOT NULL DEFAULT 0,
    ADD COLUMN paid_interest  bigint  NOT NULL DEFAULT 0,
    ADD COLUMN early          boolean NOT NULL DEFAULT false,
    ADD CONSTRAINT payments_paid CHECK (paid_principal BETWEEN 0 AND principal AND paid_interest BETWEEN 0 AND interest);

-- Строки, заменённые перерасчётом: schedule_version — версия графика, в которой они действовали.
CREATE TABLE servicing.payments_archive (
    loan_id          text   NOT NULL REFERENCES servicing.loans (id) ON DELETE CASCADE,
    schedule_version int    NOT NULL,
    n                int    NOT NULL,
    due_date         date   NOT NULL,
    principal        bigint NOT NULL,
    interest         bigint NOT NULL,
    total            bigint NOT NULL,
    paid_principal   bigint NOT NULL,
    paid_interest    bigint NOT NULL,
    PRIMARY KEY (loan_id, schedule_version, n)
);

CREATE SEQUENCE servicing.quote_seq;

CREATE TABLE servicing.quotes (
    id               text PRIMARY KEY,
    loan_id          text        NOT NULL REFERENCES servicing.loans (id) ON DELETE CASCADE,
    type             text        NOT NULL CHECK (type IN ('PARTIAL', 'FULL')),
    strategy         text CHECK (strategy IN ('REDUCE_TERM', 'REDUCE_PAYMENT')),
    amount_to_debit  bigint      NOT NULL CHECK (amount_to_debit > 0),
    alloc_principal  bigint      NOT NULL,
    alloc_interest   bigint      NOT NULL,
    interest_credit  bigint      NOT NULL DEFAULT 0, -- проценты, оплаченные заранее сверх начисленных: идут в основной долг
    monthly_after    bigint      NOT NULL,
    preview          jsonb       NOT NULL,
    loan_version     int         NOT NULL,
    schedule_version int         NOT NULL,
    effective_date   date        NOT NULL,
    calculated_at    timestamptz NOT NULL,
    valid_until      timestamptz NOT NULL,
    used_at          timestamptz,
    idem_key         text        NOT NULL,
    fingerprint      text        NOT NULL,
    UNIQUE (loan_id, idem_key)
);

CREATE SEQUENCE servicing.operation_seq;

CREATE TABLE servicing.operations (
    id                     text PRIMARY KEY,
    loan_id                text        NOT NULL REFERENCES servicing.loans (id) ON DELETE CASCADE,
    type                   text        NOT NULL CHECK (type IN ('REGULAR', 'EARLY')),
    status                 text        NOT NULL CHECK (status IN ('PENDING', 'PROCESSING', 'SUCCEEDED', 'FAILED')),
    account_id             text        NOT NULL,
    account_number         text        NOT NULL,
    requested_amount       bigint      NOT NULL CHECK (requested_amount > 0),
    quote_id               text REFERENCES servicing.quotes (id),
    idem_key               text        NOT NULL,
    fingerprint            text        NOT NULL,
    external_id            text UNIQUE,
    posted_amount          bigint,
    posted_at              timestamptz,
    failure_reason         text,
    alloc_principal        bigint,
    alloc_interest         bigint,
    alloc_penalties        bigint,
    alloc_fees             bigint,
    unapplied              bigint,
    loan_version_after     int         NOT NULL,
    schedule_version_after int         NOT NULL,
    loan_status_after      text        NOT NULL,
    requested_at           timestamptz NOT NULL,
    updated_at             timestamptz NOT NULL,
    UNIQUE (loan_id, idem_key)
);

CREATE INDEX operations_loan ON servicing.operations (loan_id, requested_at DESC, id DESC);
CREATE INDEX operations_open ON servicing.operations (status) WHERE status IN ('PENDING', 'PROCESSING');
-- Одна незавершённая денежная операция на кредит: вторая не пройдёт и при гонке.
CREATE UNIQUE INDEX operations_one_open ON servicing.operations (loan_id) WHERE status IN ('PENDING', 'PROCESSING');

-- Проводки: из них восстанавливаются даты оплат строк для пеней. n — строка графика; у пеней n нет. Перенос
-- переплаченных процентов в основной долг — пара проводок с минусом и плюсом, а не правка прежних.
CREATE TABLE servicing.postings (
    id        bigserial PRIMARY KEY,
    op_id     text        NOT NULL REFERENCES servicing.operations (id) ON DELETE CASCADE,
    loan_id   text        NOT NULL REFERENCES servicing.loans (id) ON DELETE CASCADE,
    n         int,
    component text        NOT NULL CHECK (component IN ('PRINCIPAL', 'INTEREST', 'PENALTY')),
    amount    bigint      NOT NULL CHECK (amount <> 0),
    posted_at timestamptz NOT NULL,
    CHECK ((component = 'PENALTY') = (n IS NULL))
);

CREATE INDEX postings_loan ON servicing.postings (loan_id, posted_at);

-- +goose Down
DROP TABLE servicing.postings;
DROP TABLE servicing.operations;
DROP SEQUENCE servicing.operation_seq;
DROP TABLE servicing.quotes;
DROP SEQUENCE servicing.quote_seq;
DROP TABLE servicing.payments_archive;
ALTER TABLE servicing.payments
    DROP CONSTRAINT payments_paid,
    DROP COLUMN early,
    DROP COLUMN paid_interest,
    DROP COLUMN paid_principal;
ALTER TABLE servicing.loans
    DROP CONSTRAINT loans_closed,
    DROP COLUMN close_reason,
    DROP COLUMN closed_at,
    DROP COLUMN penalties_paid,
    DROP COLUMN revolving,
    DROP COLUMN schedule_version,
    DROP COLUMN version;
