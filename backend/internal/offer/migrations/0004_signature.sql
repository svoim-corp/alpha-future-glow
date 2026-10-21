-- +goose Up
-- Имитация подписи договора: время и идентификатор подписи SIM-XXXX.
ALTER TABLE offer.offers
    ADD COLUMN signed_at    timestamptz,
    ADD COLUMN signature_id text,
    ADD CONSTRAINT offers_signature CHECK ((signed_at IS NULL) = (signature_id IS NULL)),
    ADD CONSTRAINT offers_signed_accepted CHECK (signed_at IS NULL OR accepted_at IS NOT NULL);

-- +goose Down
ALTER TABLE offer.offers DROP CONSTRAINT offers_signed_accepted, DROP CONSTRAINT offers_signature,
    DROP COLUMN signature_id, DROP COLUMN signed_at;
