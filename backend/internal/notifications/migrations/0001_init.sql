-- +goose Up
-- Уведомления в кабинете (docs/notifications.md): замена SMS и почте, всё внутри продукта.
CREATE SEQUENCE notifications.notification_seq;

CREATE TABLE notifications.notifications (
    id         text PRIMARY KEY,
    user_id    text        NOT NULL,
    type       text        NOT NULL,
    title      text        NOT NULL,
    text       text        NOT NULL,
    link       text        NOT NULL,
    read       boolean     NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL
);

CREATE INDEX notifications_user ON notifications.notifications (user_id, created_at DESC, id);

-- +goose Down
DROP TABLE notifications.notifications;
DROP SEQUENCE notifications.notification_seq;
