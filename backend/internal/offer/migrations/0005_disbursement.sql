-- +goose Up
-- Имитация выдачи: время выдачи и кредит в Servicing.
ALTER TABLE offer.offers
    ADD COLUMN disbursed_at timestamptz,
    ADD COLUMN loan_id      text,
    ADD CONSTRAINT offers_disbursement CHECK ((disbursed_at IS NULL) = (loan_id IS NULL)),
    ADD CONSTRAINT offers_disbursed_signed CHECK (disbursed_at IS NULL OR signed_at IS NOT NULL);

-- +goose Down
ALTER TABLE offer.offers DROP CONSTRAINT offers_disbursed_signed, DROP CONSTRAINT offers_disbursement,
    DROP COLUMN loan_id, DROP COLUMN disbursed_at;
