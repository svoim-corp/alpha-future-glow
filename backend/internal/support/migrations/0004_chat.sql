-- +goose Up
CREATE TABLE support.conversations (
 id text PRIMARY KEY, user_id text NOT NULL, company_id text NOT NULL,
 application_id text, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX conversations_user ON support.conversations(user_id, created_at DESC, id);
CREATE TABLE support.messages (
 id text PRIMARY KEY, seq bigserial UNIQUE NOT NULL,
 conversation_id text NOT NULL REFERENCES support.conversations(id) ON DELETE CASCADE,
 role text NOT NULL CHECK (role IN ('user','assistant')),
 status text NOT NULL CHECK (status IN ('PENDING','COMPLETED','FAILED')),
 text text NOT NULL DEFAULT '', payload jsonb NOT NULL DEFAULT '{"sources":[],"blocks":[]}',
 request_key text, created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(conversation_id, request_key)
);
CREATE UNIQUE INDEX one_pending_message ON support.messages(conversation_id) WHERE status='PENDING';
CREATE INDEX messages_conversation ON support.messages(conversation_id, seq);
CREATE TABLE support.actions (
 id text PRIMARY KEY, conversation_id text NOT NULL REFERENCES support.conversations(id) ON DELETE CASCADE,
 params jsonb NOT NULL, result jsonb, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE support.knowledge_chunks (
 id text PRIMARY KEY, title text NOT NULL, href text NOT NULL, version text NOT NULL,
 hash text NOT NULL, content text NOT NULL, embedding real[] NOT NULL,
 search tsvector GENERATED ALWAYS AS (to_tsvector('russian', content)) STORED
);
CREATE INDEX knowledge_search ON support.knowledge_chunks USING gin(search);

-- +goose Down
DROP TABLE support.knowledge_chunks;
DROP TABLE support.actions;
DROP TABLE support.messages;
DROP TABLE support.conversations;
