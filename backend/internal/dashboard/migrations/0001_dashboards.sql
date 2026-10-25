-- +goose Up
-- Настройка главной клиента (docs/dashboard.md): одна запись на пользователя в компании.
-- В config — только типы, координаты, ID источников и фильтры, без сумм и пользовательских текстов.
CREATE TABLE dashboard.dashboards (
    user_id    text        NOT NULL,
    company_id text        NOT NULL,
    revision   text        NOT NULL,
    config     jsonb       NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (user_id, company_id)
);

-- Успешные сохранения по Idempotency-Key: повтор того же тела возвращает исходный результат,
-- даже если после него настройку уже сохранили снова.
CREATE TABLE dashboard.save_keys (
    user_id    text        NOT NULL,
    company_id text        NOT NULL,
    key        text        NOT NULL,
    body_hash  text        NOT NULL,
    revision   text        NOT NULL,
    config     jsonb       NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (user_id, company_id, key)
);

-- +goose Down
DROP TABLE dashboard.save_keys;
DROP TABLE dashboard.dashboards;
