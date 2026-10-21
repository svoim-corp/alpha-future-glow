-- +goose Up
-- Номер заявки КЗ-<год>-<номер> присваивается при отправке; нумерация сквозная.
CREATE SEQUENCE origination.number_seq;

-- +goose Down
DROP SEQUENCE origination.number_seq;
