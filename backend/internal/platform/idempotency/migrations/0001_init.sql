-- +goose Up
CREATE TABLE platform.idempotency_keys (
    user_id      text NOT NULL,
    key          text NOT NULL,
    method       text NOT NULL,
    path         text NOT NULL,
    status       int NOT NULL,
    body         bytea,
    content_type text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, key)
);

-- +goose Down
DROP TABLE platform.idempotency_keys;
