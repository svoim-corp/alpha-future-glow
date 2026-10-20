-- +goose Up
CREATE TABLE identity.users (
    id            text PRIMARY KEY,
    login         text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    role          text NOT NULL,
    company_id    text,
    full_name     text NOT NULL,
    position      text NOT NULL
);

CREATE TABLE identity.sessions (
    token      text PRIMARY KEY,
    user_id    text NOT NULL REFERENCES identity.users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE identity.sessions;
DROP TABLE identity.users;
