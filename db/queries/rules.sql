-- name: ListRulesForUser :many
SELECT r.rule_id, r.bucket_id, r.name, r.priority,
       r.description_contains, r.min_amount_cents, r.max_amount_cents,
       r.transaction_type, r.category_id, r.created_at,
       b.name AS bucket_name
FROM float.rules r
JOIN float.buckets b USING (bucket_id)
WHERE b.user_id = $1
ORDER BY r.priority ASC, r.created_at ASC;

-- name: ListRulesByBucket :many
SELECT r.rule_id, r.bucket_id, r.name, r.priority,
       r.description_contains, r.min_amount_cents, r.max_amount_cents,
       r.transaction_type, r.category_id, r.created_at,
       b.name AS bucket_name
FROM float.rules r
JOIN float.buckets b USING (bucket_id)
WHERE r.bucket_id = $1 AND b.user_id = $2
ORDER BY r.priority ASC, r.created_at ASC;

-- name: CreateRule :one
INSERT INTO float.rules (bucket_id, name, priority, description_contains, min_amount_cents, max_amount_cents, transaction_type, category_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING rule_id, bucket_id, name, priority, description_contains, min_amount_cents, max_amount_cents, transaction_type, category_id, created_at;

-- name: UpdateRule :one
UPDATE float.rules r
SET name                 = $2,
    priority             = $3,
    description_contains = $4,
    min_amount_cents     = $5,
    max_amount_cents     = $6,
    transaction_type     = $7,
    category_id          = $8
FROM float.buckets b
WHERE r.rule_id = $1 AND r.bucket_id = b.bucket_id AND b.user_id = $9
RETURNING r.rule_id, r.bucket_id, r.name, r.priority, r.description_contains, r.min_amount_cents, r.max_amount_cents, r.transaction_type, r.category_id, r.created_at;

-- name: DeleteRule :exec
DELETE FROM float.rules r
USING float.buckets b
WHERE r.rule_id = $1 AND r.bucket_id = b.bucket_id AND b.user_id = $2;

-- name: ListUpTransactionsByBucketID :many
SELECT * FROM float.up_transactions WHERE bucket_id = $1;
