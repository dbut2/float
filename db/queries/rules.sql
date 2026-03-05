-- name: ListRulesForUser :many
SELECT r.rule_id, r.bucket_id, r.name, r.priority,
       r.description_contains, r.min_amount_cents, r.max_amount_cents,
       r.transaction_type, r.category_id,
       r.date_from, r.date_to, r.foreign_currency_code,
       r.created_at,
       b.name AS bucket_name
FROM float.rules r
JOIN float.buckets b USING (bucket_id)
WHERE b.user_id = $1
ORDER BY r.priority ASC, r.created_at ASC;

-- name: ListRulesByBucket :many
SELECT r.rule_id, r.bucket_id, r.name, r.priority,
       r.description_contains, r.min_amount_cents, r.max_amount_cents,
       r.transaction_type, r.category_id,
       r.date_from, r.date_to, r.foreign_currency_code,
       r.created_at,
       b.name AS bucket_name
FROM float.rules r
JOIN float.buckets b USING (bucket_id)
WHERE r.bucket_id = $1 AND b.user_id = $2
ORDER BY r.priority ASC, r.created_at ASC;

-- name: CreateRule :one
INSERT INTO float.rules (bucket_id, name, priority, description_contains, min_amount_cents, max_amount_cents, transaction_type, category_id, date_from, date_to, foreign_currency_code)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING rule_id, bucket_id, name, priority, description_contains, min_amount_cents, max_amount_cents, transaction_type, category_id, date_from, date_to, foreign_currency_code, created_at;

-- name: UpdateRule :one
UPDATE float.rules r
SET name                 = $2,
    priority             = $3,
    description_contains = $4,
    min_amount_cents     = $5,
    max_amount_cents     = $6,
    transaction_type     = $7,
    category_id          = $8,
    date_from            = $9,
    date_to              = $10,
    foreign_currency_code = $11
FROM float.buckets b
WHERE r.rule_id = $1 AND r.bucket_id = b.bucket_id AND b.user_id = $12
RETURNING r.rule_id, r.bucket_id, r.name, r.priority, r.description_contains, r.min_amount_cents, r.max_amount_cents, r.transaction_type, r.category_id, r.date_from, r.date_to, r.foreign_currency_code, r.created_at;

-- name: DeleteRule :exec
DELETE FROM float.rules r
USING float.buckets b
WHERE r.rule_id = $1 AND r.bucket_id = b.bucket_id AND b.user_id = $2;

-- name: ListUpTransactionsByBucketID :many
SELECT transaction_id, bucket_id, description, message, amount_cents, foreign_currency_code, foreign_amount_cents, created_at, transaction_type, raw_json, category_id FROM float.up_transactions WHERE bucket_id = $1;
