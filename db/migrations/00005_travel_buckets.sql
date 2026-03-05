-- +goose Up

DROP VIEW IF EXISTS float.bucket_ledger;

ALTER TABLE float.up_transactions
    DROP COLUMN currency_code,
    DROP COLUMN display_amount,
    ADD COLUMN foreign_currency_code TEXT,
    ADD COLUMN foreign_amount_cents  BIGINT;

ALTER TABLE float.buckets
    ADD COLUMN currency_code TEXT;

CREATE TABLE float.fx_rates (
    base_currency  TEXT NOT NULL,
    quote_currency TEXT NOT NULL,
    rate           DOUBLE PRECISION NOT NULL,
    date           DATE NOT NULL DEFAULT CURRENT_DATE,
    PRIMARY KEY (base_currency, quote_currency, date)
);

CREATE OR REPLACE VIEW float.bucket_ledger AS
    SELECT transaction_id,
           bucket_id,
           description,
           message,
           amount_cents,
           foreign_currency_code,
           foreign_amount_cents,
           created_at,
           TRUE AS is_transaction
    FROM float.up_transactions
    WHERE transaction_type IS DISTINCT FROM 'Transfer'
      AND transaction_type IS DISTINCT FROM 'Round Up'

    UNION ALL

    SELECT NULL::UUID AS transaction_id,
           to_bucket_id AS bucket_id,
           note AS description,
           '' AS message,
           amount_cents,
           NULL AS foreign_currency_code,
           NULL::BIGINT AS foreign_amount_cents,
           created_at,
           FALSE AS is_transaction
    FROM float.bucket_transfers

    UNION ALL

    SELECT NULL::UUID AS transaction_id,
           from_bucket_id AS bucket_id,
           note AS description,
           '' AS message,
           -amount_cents,
           NULL AS foreign_currency_code,
           NULL::BIGINT AS foreign_amount_cents,
           created_at,
           FALSE AS is_transaction
    FROM float.bucket_transfers;

-- +goose Down

ALTER TABLE float.up_transactions
    DROP COLUMN IF EXISTS foreign_amount_cents,
    DROP COLUMN IF EXISTS foreign_currency_code,
    ADD COLUMN currency_code  TEXT NOT NULL DEFAULT 'AUD',
    ADD COLUMN display_amount TEXT NOT NULL DEFAULT '';

CREATE OR REPLACE VIEW float.bucket_ledger AS
    SELECT transaction_id,
           bucket_id,
           description,
           message,
           amount_cents,
           display_amount,
           currency_code,
           created_at,
           TRUE AS is_transaction
    FROM float.up_transactions
    WHERE transaction_type IS DISTINCT FROM 'Transfer'
      AND transaction_type IS DISTINCT FROM 'Round Up'

    UNION ALL

    SELECT NULL as transaction_id,
           to_bucket_id AS bucket_id,
           note AS description,
           '' AS message,
           amount_cents,
           '' AS display_amount,
           'AUD' AS currency_code,
           created_at,
           FALSE AS is_transaction
    FROM float.bucket_transfers

    UNION ALL

    SELECT NULL as transaction_id,
           from_bucket_id AS bucket_id,
           note AS description,
           '' AS message,
           -amount_cents,
           '' AS display_amount,
           'AUD' AS currency_code,
           created_at,
           FALSE AS is_transaction
    FROM float.bucket_transfers;

DROP TABLE IF EXISTS float.fx_rates;
ALTER TABLE float.buckets DROP COLUMN IF EXISTS currency_code;
