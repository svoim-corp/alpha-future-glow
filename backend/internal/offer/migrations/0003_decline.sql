-- +goose Up
-- Отказ клиента от предложения: заявка закрывается, причина остаётся для банка.
ALTER TABLE offer.offers
    ADD COLUMN declined_at    timestamptz,
    ADD COLUMN decline_reason text,
    ADD CONSTRAINT offers_one_answer CHECK (accepted_at IS NULL OR declined_at IS NULL);

-- +goose Down
ALTER TABLE offer.offers DROP CONSTRAINT offers_one_answer, DROP COLUMN decline_reason, DROP COLUMN declined_at;
