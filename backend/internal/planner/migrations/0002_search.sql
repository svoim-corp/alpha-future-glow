-- +goose Up
CREATE INDEX plans_search_text_idx ON planner.plans USING gin (to_tsvector('russian', replace(lower(id || ' ' || title || ' ' || goal || ' ' || coalesce(note, '')), 'ё', 'е')));

-- +goose Down
DROP INDEX planner.plans_search_text_idx;
