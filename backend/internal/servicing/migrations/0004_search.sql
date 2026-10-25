-- +goose Up
CREATE INDEX loans_search_text_idx ON servicing.loans USING gin (to_tsvector('russian', replace(lower(id || ' ' || contract_number || ' ' || product_code || ' ' || product_name), 'ё', 'е')));

-- +goose Down
DROP INDEX servicing.loans_search_text_idx;
