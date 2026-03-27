-- name: UpsertUpTransaction :one
INSERT INTO float.up_transactions (
    transaction_id, bucket_id, description, message,
    amount_cents, created_at, transaction_type,
    raw_json, category_id, foreign_currency_code, foreign_amount_cents
) VALUES (
    $1,
    (SELECT bucket_id FROM float.buckets WHERE user_id = $2 AND is_general = TRUE),
    $3, $4, $5, $6, $7, $8, $9, $10, $11
)
ON CONFLICT (transaction_id) DO UPDATE SET
    description = EXCLUDED.description,
    message = EXCLUDED.message,
    amount_cents = EXCLUDED.amount_cents,
    created_at = EXCLUDED.created_at,
    transaction_type = EXCLUDED.transaction_type,
    raw_json = EXCLUDED.raw_json,
    category_id = EXCLUDED.category_id,
    foreign_currency_code = EXCLUDED.foreign_currency_code,
    foreign_amount_cents = EXCLUDED.foreign_amount_cents
RETURNING (xmax = 0) AS inserted;

-- name: AssignTransactionToBucket :exec
UPDATE float.up_transactions
SET bucket_id = $2
WHERE transaction_id = $1
AND bucket_id IN (SELECT bucket_id FROM float.buckets WHERE user_id = $3);

-- name: DeleteUpTransaction :exec
DELETE FROM float.up_transactions WHERE transaction_id = $1;

-- name: ListUpTransactionsByUser :many
SELECT t.* FROM float.up_transactions t
JOIN float.buckets b ON t.bucket_id = b.bucket_id
WHERE b.user_id = $1
  AND t.transaction_type IS DISTINCT FROM 'Transfer'
  AND t.transaction_type IS DISTINCT FROM 'Round Up'
ORDER BY t.created_at DESC;

-- name: ListUpTransactionsByBucket :many
SELECT t.* FROM float.up_transactions t
JOIN float.buckets b ON t.bucket_id = b.bucket_id
WHERE t.bucket_id = $1 AND b.user_id = $2
  AND t.transaction_type IS DISTINCT FROM 'Transfer'
  AND t.transaction_type IS DISTINCT FROM 'Round Up'
ORDER BY t.created_at DESC;

-- name: GetUpTransaction :one
SELECT t.* FROM float.up_transactions t
JOIN float.buckets b ON t.bucket_id = b.bucket_id
WHERE t.transaction_id = $1 AND b.user_id = $2;
