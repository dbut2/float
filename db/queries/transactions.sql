-- name: UpsertUpTransaction :one
INSERT INTO float.up_transactions (
    transaction_id, bucket_id, description, message,
    amount_cents, display_amount, currency_code, created_at, transaction_type,
    raw_json
) VALUES (
    $1,
    (SELECT bucket_id FROM float.buckets WHERE user_id = $2 AND is_general = TRUE),
    $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (transaction_id) DO UPDATE SET
    description = EXCLUDED.description,
    message = EXCLUDED.message,
    amount_cents = EXCLUDED.amount_cents,
    display_amount = EXCLUDED.display_amount,
    currency_code = EXCLUDED.currency_code,
    created_at = EXCLUDED.created_at,
    transaction_type = EXCLUDED.transaction_type,
    deep_link_url = EXCLUDED.deep_link_url,
    raw_json = EXCLUDED.raw_json
RETURNING (xmax = 0) AS inserted;

-- name: GetTransaction :one
SELECT l.* FROM float.bucket_ledger l
JOIN float.buckets b USING (bucket_id)
WHERE l.transaction_id = $1 AND b.user_id = $2;

-- name: ListTransactions :many
SELECT l.* FROM float.bucket_ledger l
JOIN float.buckets b USING (bucket_id)
WHERE b.user_id = $1
ORDER BY l.created_at DESC;

-- name: AssignTransactionToBucket :exec
UPDATE float.up_transactions
SET bucket_id = $2
WHERE transaction_id = $1;

-- name: ListBucketTransactions :many
SELECT * FROM float.bucket_ledger
WHERE bucket_id = $1
ORDER BY created_at DESC;

-- name: DeleteUpTransaction :exec
DELETE FROM float.up_transactions WHERE transaction_id = $1;
