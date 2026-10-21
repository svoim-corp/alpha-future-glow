-- +goose Up
-- Кэш профиля клиента из АБС и реестров: GET /applications/{id} и разбор заявки читают профиль часто, а моки — внешние сервисы.
CREATE TABLE company.profiles (
    company_id text PRIMARY KEY REFERENCES company.companies (id),
    profile    jsonb NOT NULL,
    fetched_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE company.profiles;
