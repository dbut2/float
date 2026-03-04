-- +goose Up

ALTER TABLE float.up_transactions ADD COLUMN category_id TEXT;

CREATE TABLE float.rules (
    rule_id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bucket_id            UUID NOT NULL REFERENCES float.buckets(bucket_id) ON DELETE CASCADE,
    name                 TEXT NOT NULL,
    priority             INT NOT NULL DEFAULT 0,
    description_contains TEXT,
    min_amount_cents     BIGINT,
    max_amount_cents     BIGINT,
    transaction_type     TEXT,
    category_id          TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE IF EXISTS float.rules;
ALTER TABLE float.up_transactions DROP COLUMN IF EXISTS category_id;
