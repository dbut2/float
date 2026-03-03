-- +goose Up
CREATE SCHEMA IF NOT EXISTS bank;

CREATE TABLE IF NOT EXISTS bank.users (
    user_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bank.user_tokens (
    user_id     UUID PRIMARY KEY REFERENCES bank.users (user_id) ON DELETE CASCADE,
    up_token    TEXT NOT NULL,
    verified_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS bank.buckets (
    bucket_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES bank.users (user_id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    is_general BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);

CREATE TABLE IF NOT EXISTS bank.up_transactions (
    transaction_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bucket_id        UUID NOT NULL REFERENCES bank.buckets (bucket_id) ON DELETE CASCADE,
    description      TEXT NOT NULL,
    message          TEXT NOT NULL,
    amount_cents     BIGINT NOT NULL,
    display_amount   TEXT NOT NULL,
    currency_code    TEXT NOT NULL DEFAULT 'AUD',
    created_at       TIMESTAMPTZ NOT NULL,
    transaction_type TEXT,
    deep_link_url    TEXT NOT NULL,
    raw_json         JSONB
);

CREATE TABLE IF NOT EXISTS bank.fcm_tokens (
    user_id   UUID NOT NULL REFERENCES bank.users (user_id) ON DELETE CASCADE,
    fcm_token TEXT NOT NULL UNIQUE,
    PRIMARY KEY (user_id, fcm_token)
);

CREATE TABLE IF NOT EXISTS bank.bucket_transfers (
    transfer_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_bucket_id UUID NOT NULL REFERENCES bank.buckets (bucket_id) ON DELETE CASCADE,
    to_bucket_id   UUID NOT NULL REFERENCES bank.buckets (bucket_id) ON DELETE CASCADE,
    amount_cents   BIGINT NOT NULL CHECK (amount_cents > 0),
    note           TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (from_bucket_id IS DISTINCT FROM to_bucket_id)
);

CREATE OR REPLACE VIEW bank.bucket_ledger AS
    SELECT bucket_id,
           amount_cents,
           created_at
    FROM bank.up_transactions
    WHERE transaction_type IS DISTINCT FROM 'Transfer'
      AND transaction_type IS DISTINCT FROM 'Round Up'

    UNION ALL

    SELECT to_bucket_id AS bucket_id,
           amount_cents,
           created_at
    FROM bank.bucket_transfers

    UNION ALL

    SELECT from_bucket_id AS bucket_id,
           -amount_cents AS amount_cents,
           created_at
    FROM bank.bucket_transfers;

-- +goose Down
DROP VIEW IF EXISTS bank.bucket_ledger;
DROP TABLE IF EXISTS bank.bucket_transfers;
DROP TABLE IF EXISTS bank.fcm_tokens;
DROP TABLE IF EXISTS bank.up_transactions;
DROP TABLE IF EXISTS bank.buckets;
DROP TABLE IF EXISTS bank.user_tokens;
DROP TABLE IF EXISTS bank.users;
DROP SCHEMA IF EXISTS bank;
