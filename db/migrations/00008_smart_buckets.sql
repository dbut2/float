-- +goose Up

-- Natural-language description of what belongs in each bucket
ALTER TABLE float.buckets ADD COLUMN description TEXT NOT NULL DEFAULT '';

-- Classification audit log
CREATE TABLE float.classification_log (
    log_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id   UUID NOT NULL REFERENCES float.up_transactions(transaction_id) ON DELETE CASCADE,
    chosen_bucket_id UUID NOT NULL REFERENCES float.buckets(bucket_id) ON DELETE CASCADE,
    confidence       REAL NOT NULL,
    reasoning        TEXT NOT NULL DEFAULT '',
    model            TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Drop legacy rules
DROP TABLE IF EXISTS float.rules;

-- +goose Down
CREATE TABLE float.rules (
    rule_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bucket_id UUID NOT NULL REFERENCES float.buckets(bucket_id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    priority INT NOT NULL DEFAULT 0,
    description_contains TEXT,
    min_amount_cents BIGINT,
    max_amount_cents BIGINT,
    transaction_type TEXT,
    category_id TEXT,
    date_from DATE,
    date_to DATE,
    foreign_currency_code TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

DROP TABLE IF EXISTS float.classification_log;
ALTER TABLE float.buckets DROP COLUMN IF EXISTS description;
