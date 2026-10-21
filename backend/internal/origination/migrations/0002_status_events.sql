-- +goose Up
-- История статусов заявки (StatusEvent): кто, когда и каким действием перевёл заявку. Пишется в той же транзакции,
-- что и смена статуса.
CREATE TABLE origination.status_events (
    id          bigserial PRIMARY KEY,
    app_id      text        NOT NULL REFERENCES origination.applications (id) ON DELETE CASCADE,
    from_status text        NOT NULL,
    to_status   text        NOT NULL,
    action      text        NOT NULL,
    actor_type  text        NOT NULL CHECK (actor_type IN ('CLIENT', 'OFFICER', 'SYSTEM')),
    actor_id    text,
    actor_name  text,
    comment     text,
    at          timestamptz NOT NULL
);

CREATE INDEX status_events_app ON origination.status_events (app_id, at, id);

-- +goose Down
DROP TABLE origination.status_events;
