-- +goose Up
-- Согласия на запрос в БКИ: кто, когда, по какой заявке. По одному на заявителя и поручителя; id передаётся
-- в мок БКИ как consentId.
CREATE SEQUENCE origination.consent_seq;

CREATE TABLE origination.consents (
    id        text PRIMARY KEY,
    app_id    text        NOT NULL REFERENCES origination.applications (id) ON DELETE CASCADE,
    subject   text        NOT NULL CHECK (subject IN ('APPLICANT', 'GUARANTOR')),
    person_id text, -- поручитель, за которого дано согласие; у заявителя — NULL
    text      text        NOT NULL,
    given_by  text        NOT NULL,
    given_at  timestamptz NOT NULL,
    UNIQUE (app_id, subject),
    CHECK ((subject = 'GUARANTOR') = (person_id IS NOT NULL))
);

-- +goose Down
DROP TABLE origination.consents;
DROP SEQUENCE origination.consent_seq;
