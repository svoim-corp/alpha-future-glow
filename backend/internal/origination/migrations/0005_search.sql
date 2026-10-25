-- +goose Up
CREATE INDEX applications_search_text_idx ON origination.applications USING gin (to_tsvector('russian', replace(lower(id || ' ' || coalesce(number, '') || ' ' || product_code || ' ' || coalesce(goal, '') || ' ' || coalesce(purpose_text, '')), 'ё', 'е')));

-- +goose Down
DROP INDEX origination.applications_search_text_idx;
