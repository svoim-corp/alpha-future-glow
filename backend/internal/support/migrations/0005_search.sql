-- +goose Up
-- В старом main версия 4 создавала поиск, в feature/support — чат.
-- Доводим обе базы до общей схемы, не меняя уже применённую 0004_chat.sql.
CREATE TABLE IF NOT EXISTS support.conversations (
 id text PRIMARY KEY, user_id text NOT NULL, company_id text NOT NULL,
 application_id text, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS conversations_user ON support.conversations(user_id, created_at DESC, id);
CREATE TABLE IF NOT EXISTS support.messages (
 id text PRIMARY KEY, seq bigserial UNIQUE NOT NULL,
 conversation_id text NOT NULL REFERENCES support.conversations(id) ON DELETE CASCADE,
 role text NOT NULL CHECK (role IN ('user','assistant')),
 status text NOT NULL CHECK (status IN ('PENDING','COMPLETED','FAILED')),
 text text NOT NULL DEFAULT '', payload jsonb NOT NULL DEFAULT '{"sources":[],"blocks":[]}',
 request_key text, created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(conversation_id, request_key)
);
CREATE UNIQUE INDEX IF NOT EXISTS one_pending_message ON support.messages(conversation_id) WHERE status='PENDING';
CREATE INDEX IF NOT EXISTS messages_conversation ON support.messages(conversation_id, seq);
CREATE TABLE IF NOT EXISTS support.actions (
 id text PRIMARY KEY, conversation_id text NOT NULL REFERENCES support.conversations(id) ON DELETE CASCADE,
 params jsonb NOT NULL, result jsonb, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS support.knowledge_chunks (
 id text PRIMARY KEY, title text NOT NULL, href text NOT NULL, version text NOT NULL,
 hash text NOT NULL, content text NOT NULL, embedding real[] NOT NULL,
 search tsvector GENERATED ALWAYS AS (to_tsvector('russian', content)) STORED
);
CREATE INDEX IF NOT EXISTS knowledge_search ON support.knowledge_chunks USING gin(search);
CREATE INDEX IF NOT EXISTS faq_search_text_idx ON support.faq USING gin (to_tsvector('russian', replace(lower(id || ' ' || question || ' ' || answer), 'ё', 'е')));
CREATE INDEX IF NOT EXISTS tickets_search_text_idx ON support.tickets USING gin (to_tsvector('russian', replace(lower(id || ' ' || subject || ' ' || text || ' ' || coalesce(reply, '')), 'ё', 'е')));

-- +goose Down
-- Таблицы и переписки не удаляем: они принадлежат исторической версии 4.
DROP INDEX IF EXISTS support.tickets_search_text_idx;
DROP INDEX IF EXISTS support.faq_search_text_idx;
