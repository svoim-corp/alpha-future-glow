-- +goose Up
CREATE TABLE company.companies (
    id             text PRIMARY KEY,
    inn            text NOT NULL UNIQUE,
    name           text NOT NULL,
    type           text NOT NULL,
    is_bank_client boolean NOT NULL DEFAULT false
);

INSERT INTO company.companies (id, inn, name, type, is_bank_client)
VALUES ('c-1', '7712345678', 'ООО «Северная пекарня»', 'LLC', true);

-- +goose Down
DROP TABLE company.companies;
