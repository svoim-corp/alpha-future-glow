-- +goose Up
CREATE TABLE planner.assistant_commands (
    user_id text NOT NULL,
    command_id text NOT NULL,
    fingerprint text NOT NULL,
    result jsonb NOT NULL,
    PRIMARY KEY (user_id, command_id)
);

-- +goose Down
DROP TABLE planner.assistant_commands;
