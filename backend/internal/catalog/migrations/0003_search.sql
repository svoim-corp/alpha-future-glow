-- +goose Up
CREATE INDEX products_search_text_idx ON catalog.products USING gin (to_tsvector('russian', replace(lower(code || ' ' || name || ' ' || short_text || ' ' || description), 'ё', 'е')));

-- +goose Down
DROP INDEX catalog.products_search_text_idx;
