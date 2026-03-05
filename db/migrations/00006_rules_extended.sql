-- +goose Up

ALTER TABLE float.rules
    ADD COLUMN date_from             DATE,
    ADD COLUMN date_to               DATE,
    ADD COLUMN foreign_currency_code TEXT;

-- +goose Down

ALTER TABLE float.rules
    DROP COLUMN IF EXISTS foreign_currency_code,
    DROP COLUMN IF EXISTS date_to,
    DROP COLUMN IF EXISTS date_from;
