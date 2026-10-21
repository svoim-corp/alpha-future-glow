-- +goose Up
-- Напоминание о платеже отправляется один раз: время отправки хранится в строке графика.
ALTER TABLE servicing.payments ADD COLUMN reminded_at timestamptz;

-- +goose Down
ALTER TABLE servicing.payments DROP COLUMN reminded_at;
