-- +goose Up
-- Линейка продуктов и цели кредита (docs/catalog.md).
CREATE TABLE catalog.goals (
    code text PRIMARY KEY,
    name text NOT NULL,
    sort int  NOT NULL
);

CREATE TABLE catalog.products (
    code                    text PRIMARY KEY,
    name                    text          NOT NULL,
    short_text              text          NOT NULL,
    description             text          NOT NULL,
    goals                   jsonb         NOT NULL,
    amount_min              bigint        NOT NULL CHECK (amount_min > 0),
    amount_max              bigint        NOT NULL CHECK (amount_max >= amount_min),
    term_min                int           NOT NULL CHECK (term_min > 0),
    term_max                int           NOT NULL CHECK (term_max >= term_min),
    rate_min                numeric(5, 2) NOT NULL CHECK (rate_min > 0),
    rate_max                numeric(5, 2) NOT NULL CHECK (rate_max >= rate_min),
    grace_max_months        int           NOT NULL DEFAULT 0 CHECK (grace_max_months >= 0),
    min_business_age_months int           NOT NULL DEFAULT 0 CHECK (min_business_age_months >= 0),
    guarantor_for_llc       boolean       NOT NULL DEFAULT false,
    revolving               boolean       NOT NULL DEFAULT false,
    limit_rule              text          NOT NULL,
    what_you_need           jsonb         NOT NULL,
    sort                    int           NOT NULL
);

-- +goose Down
DROP TABLE catalog.products;
DROP TABLE catalog.goals;
