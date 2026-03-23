-- name: CreateBucket :one
INSERT INTO float.buckets (user_id, name, currency_code, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: SeedBucket :one
INSERT INTO float.buckets (bucket_id, user_id, name, is_general, currency_code, description)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, name) DO UPDATE SET
    bucket_id = EXCLUDED.bucket_id,
    is_general = EXCLUDED.is_general,
    currency_code = EXCLUDED.currency_code,
    description = EXCLUDED.description
RETURNING *;

-- name: ListBuckets :many
SELECT b.*, COALESCE(SUM(l.amount_cents), 0)::BIGINT AS balance_cents
FROM float.buckets b
LEFT JOIN float.bucket_ledger l ON b.bucket_id = l.bucket_id
WHERE b.user_id = $1 AND b.status = 'active'
GROUP BY b.bucket_id
ORDER BY b.display_order NULLS LAST, b.created_at ASC;

-- name: SetBucketDisplayOrder :exec
UPDATE float.buckets
SET display_order = $2
WHERE bucket_id = $1 AND user_id = $3;

-- name: GetBucket :one
SELECT b.*, COALESCE(SUM(l.amount_cents), 0)::BIGINT AS balance_cents
FROM float.buckets b
LEFT JOIN float.bucket_ledger l ON b.bucket_id = l.bucket_id
WHERE b.bucket_id = $1 AND b.user_id = $2 AND b.status = 'active'
GROUP BY b.bucket_id;

-- name: EnsureGeneralBucket :exec
INSERT INTO float.buckets (user_id, name, is_general)
VALUES ($1, 'General', TRUE)
ON CONFLICT (user_id, name) DO UPDATE SET is_general = TRUE;

-- name: DeleteBucket :exec
DELETE FROM float.buckets
WHERE bucket_id = $1 AND user_id = $2;

-- name: ReassignBucketTransactionsToGeneral :exec
UPDATE float.up_transactions t
SET bucket_id = (SELECT b2.bucket_id FROM float.buckets b1 JOIN float.buckets b2 ON b1.user_id = b2.user_id AND b2.is_general = TRUE WHERE b1.bucket_id = $1)
WHERE t.bucket_id = $1;

-- name: CloseBucket :exec
UPDATE float.buckets SET status = 'closed' WHERE bucket_id = $1 AND user_id = $2 AND status = 'active';

-- name: GetGeneralBucket :one
SELECT * FROM float.buckets WHERE user_id = $1 AND is_general = TRUE;

-- name: UpdateBucketDescription :exec
UPDATE float.buckets SET description = $2 WHERE bucket_id = $1 AND user_id = $3;

-- name: ListBucketSampleTransactions :many
SELECT DISTINCT ON (t.description) t.description, t.amount_cents, t.category_id, t.foreign_currency_code
FROM float.up_transactions t
WHERE t.bucket_id = $1
ORDER BY t.description, t.created_at DESC
LIMIT 10;
