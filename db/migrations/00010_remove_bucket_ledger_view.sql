-- +goose Up
DROP VIEW IF EXISTS float.bucket_ledger;

-- +goose Down
CREATE OR REPLACE VIEW float.bucket_ledger AS
    SELECT transaction_id,
           bucket_id,
           description,
           message,
           amount_cents,
           foreign_currency_code,
           foreign_amount_cents,
           created_at,
           TRUE AS is_transaction,
           NULL::UUID AS covers_transaction_id
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
           FALSE AS is_transaction,
           covers_transaction_id
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
           FALSE AS is_transaction,
           covers_transaction_id
    FROM float.bucket_transfers;
