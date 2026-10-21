-- +goose Up
-- Заявки (docs/origination.md). Полномочия заявителя — снимок на момент создания черновика: по нему потом проверяются
-- согласие на БКИ и права в мастере, даже если данные банка изменятся.
CREATE SEQUENCE origination.application_seq START 1001;

CREATE TABLE origination.applications (
    id                  text PRIMARY KEY,
    number              text UNIQUE,
    company_id          text        NOT NULL,
    applicant_id        text        NOT NULL,
    product_code        text        NOT NULL,
    goal                text,
    purpose_text        text,
    amount              bigint      NOT NULL CHECK (amount > 0),
    term                int         NOT NULL CHECK (term > 0),
    grace_months        int CHECK (grace_months > 0),
    overdraft_limit     bigint CHECK (overdraft_limit > 0),
    guarantor_person_id text,
    truth_confirmed     boolean     NOT NULL DEFAULT false,
    authority           jsonb       NOT NULL,
    status              text        NOT NULL CHECK (status IN ('DRAFT', 'SUBMITTED', 'DATA_CHECK', 'INFO_REQUESTED', 'SCORING',
        'ON_DECISION', 'APPROVED', 'COUNTER_OFFER', 'REJECTED', 'SIGNING', 'ISSUED', 'CLOSED')),
    version             int         NOT NULL DEFAULT 1,
    created_at          timestamptz NOT NULL,
    updated_at          timestamptz NOT NULL,
    submitted_at        timestamptz,
    deleted_at          timestamptz,
    CHECK ((number IS NULL) = (submitted_at IS NULL))
);

CREATE INDEX applications_company ON origination.applications (company_id, updated_at DESC, id);

-- +goose Down
DROP TABLE origination.applications;
DROP SEQUENCE origination.application_seq;
