-- +goose Up
CREATE INDEX faq_search_text_idx ON support.faq USING gin (to_tsvector('russian', replace(lower(id || ' ' || question || ' ' || answer), 'ё', 'е')));
CREATE INDEX tickets_search_text_idx ON support.tickets USING gin (to_tsvector('russian', replace(lower(id || ' ' || subject || ' ' || text || ' ' || coalesce(reply, '')), 'ё', 'е')));

-- +goose Down
DROP INDEX support.tickets_search_text_idx;
DROP INDEX support.faq_search_text_idx;
