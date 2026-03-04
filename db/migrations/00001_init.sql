-- +goose Up
CREATE SCHEMA IF NOT EXISTS float;

CREATE TABLE IF NOT EXISTS float.users (
    user_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email          TEXT NOT NULL UNIQUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    up_token       TEXT,
    webhook_secret TEXT
);

CREATE TABLE IF NOT EXISTS float.buckets (
    bucket_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES float.users (user_id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    is_general BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);

CREATE TABLE IF NOT EXISTS float.up_transactions (
    transaction_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bucket_id        UUID NOT NULL REFERENCES float.buckets (bucket_id) ON DELETE CASCADE,
    description      TEXT NOT NULL,
    message          TEXT NOT NULL,
    amount_cents     BIGINT NOT NULL,
    display_amount   TEXT NOT NULL,
    currency_code    TEXT NOT NULL DEFAULT 'AUD',
    created_at       TIMESTAMPTZ NOT NULL,
    transaction_type TEXT,
    deep_link_url    TEXT NOT NULL,
    raw_json         JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS float.fcm_tokens (
    user_id   UUID NOT NULL REFERENCES float.users (user_id) ON DELETE CASCADE,
    fcm_token TEXT NOT NULL UNIQUE,
    PRIMARY KEY (user_id, fcm_token)
);

CREATE TABLE IF NOT EXISTS float.bucket_transfers (
    transfer_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_bucket_id UUID NOT NULL REFERENCES float.buckets (bucket_id) ON DELETE CASCADE,
    to_bucket_id   UUID NOT NULL REFERENCES float.buckets (bucket_id) ON DELETE CASCADE,
    amount_cents   BIGINT NOT NULL CHECK (amount_cents > 0),
    note           TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (from_bucket_id IS DISTINCT FROM to_bucket_id)
);

CREATE OR REPLACE VIEW float.bucket_ledger AS
    SELECT transaction_id,
           bucket_id,
           description,
           message,
           amount_cents,
           display_amount,
           currency_code,
           created_at,
           deep_link_url,
           TRUE AS is_transaction
    FROM float.up_transactions
    WHERE transaction_type IS DISTINCT FROM 'Transfer'
      AND transaction_type IS DISTINCT FROM 'Round Up'

    UNION ALL

    SELECT NULL as transaction_id,
           to_bucket_id AS bucket_id,
           note AS description,
           amount_cents,
           NULL AS display_amount, --todo
           'AUD' AS currency_code,
           created_at,
           FALSE AS is_transaction
    FROM float.bucket_transfers

    UNION ALL

    SELECT NULL as transaction_id,
           from_bucket_id AS bucket_id,
           note AS description,
           -amount_cents,
           NULL AS display_amount, --todo
           'AUD' AS currency_code,
           created_at,
           FALSE AS is_transaction
    FROM float.bucket_transfers;

-- +goose Down
DROP VIEW IF EXISTS float.bucket_ledger;
DROP TABLE IF EXISTS float.bucket_transfers;
DROP TABLE IF EXISTS float.fcm_tokens;
DROP TABLE IF EXISTS float.up_transactions;
DROP TABLE IF EXISTS float.buckets;
DROP TABLE IF EXISTS float.users;
DROP SCHEMA IF EXISTS float;
