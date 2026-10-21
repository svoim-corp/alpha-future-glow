-- +goose Up
-- Планы развития компании (BusinessPlan, API_SPEC_ADDITIONAL.md). Рекомендация не хранится — считается при чтении.
CREATE SEQUENCE planner.plan_seq;

CREATE TABLE planner.plans (
    id            text PRIMARY KEY,
    company_id    text NOT NULL,
    created_by    text NOT NULL,
    title         text NOT NULL,
    note          text,
    goal          text NOT NULL,
    budget_rub    bigint NOT NULL CHECK (budget_rub > 0),
    date_mode     text NOT NULL CHECK (date_mode IN ('DAY', 'PERIOD', 'UNSCHEDULED')),
    start_date    date,
    end_date      date,
    state         text NOT NULL CHECK (state IN ('ACTIVE', 'PAUSED', 'DONE', 'ARCHIVED')),
    notify_in_app boolean NOT NULL,
    notify_email  boolean NOT NULL,
    created_at    timestamptz NOT NULL,
    updated_at    timestamptz NOT NULL,
    CHECK ((date_mode = 'UNSCHEDULED') = (start_date IS NULL)),
    CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX plans_company ON planner.plans (company_id, created_at DESC, id);

-- Какие варианты уже предложены по плану: PLAN_MATCH — не чаще одного раза на вариант.
CREATE TABLE planner.plan_matches (
    plan_id      text NOT NULL REFERENCES planner.plans (id) ON DELETE CASCADE,
    product_code text NOT NULL,
    notified_at  timestamptz NOT NULL,
    PRIMARY KEY (plan_id, product_code)
);

-- +goose Down
DROP TABLE planner.plan_matches;
DROP TABLE planner.plans;
DROP SEQUENCE planner.plan_seq;
