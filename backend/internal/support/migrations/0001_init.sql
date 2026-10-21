-- +goose Up
-- Частые вопросы по шагам пути клиента (docs/support.md).
CREATE TABLE support.faq (
    id       text PRIMARY KEY,
    step     text NOT NULL CHECK (step IN ('PRODUCT', 'APPLICATION', 'DOCUMENTS', 'STATUS', 'DECISION')),
    question text NOT NULL,
    answer   text NOT NULL,
    sort     int  NOT NULL
);

-- +goose Down
DROP TABLE support.faq;
