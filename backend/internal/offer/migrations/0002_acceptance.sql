-- +goose Up
-- Принятие предложения: с ним появляется договор с номером КД-<год>-<номер>.
CREATE SEQUENCE offer.contract_seq;

ALTER TABLE offer.offers
    ADD COLUMN accepted_at     timestamptz,
    ADD COLUMN contract_number text UNIQUE,
    ADD CONSTRAINT offers_contract CHECK ((accepted_at IS NULL) = (contract_number IS NULL));

-- +goose Down
ALTER TABLE offer.offers DROP CONSTRAINT offers_contract, DROP COLUMN contract_number, DROP COLUMN accepted_at;
DROP SEQUENCE offer.contract_seq;
