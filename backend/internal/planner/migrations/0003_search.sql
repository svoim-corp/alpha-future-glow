-- +goose Up
-- Объединяет базы со старой 0002_search.sql и с 0002_assistant_commands.sql.
CREATE TABLE IF NOT EXISTS planner.assistant_commands (
    user_id text NOT NULL,
    command_id text NOT NULL,
    fingerprint text NOT NULL,
    result jsonb NOT NULL,
    PRIMARY KEY (user_id, command_id)
);
CREATE INDEX IF NOT EXISTS plans_search_text_idx ON planner.plans USING gin (to_tsvector('russian', replace(lower(id || ' ' || title || ' ' || goal || ' ' || coalesce(note, '')), 'ё', 'е')));

-- +goose Down
-- Подтверждённые команды сохраняем: их результат защищает от повторного изменения.
DROP INDEX IF EXISTS planner.plans_search_text_idx;
