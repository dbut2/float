-- name: CreateBucket :one
INSERT INTO float.buckets (user_id, name)
VALUES ($1, $2)
RETURNING *;

-- name: ListBuckets :many
SELECT b.*, COALESCE(SUM(l.amount_cents), 0)::BIGINT AS balance_cents
FROM float.buckets b
LEFT JOIN float.bucket_ledger l ON b.bucket_id = l.bucket_id
WHERE b.user_id = $1
GROUP BY b.bucket_id
ORDER BY b.name ASC;

-- name: GetBucket :one
SELECT b.*, COALESCE(SUM(l.amount_cents), 0)::BIGINT AS balance_cents
FROM float.buckets b
LEFT JOIN float.bucket_ledger l ON b.bucket_id = l.bucket_id
WHERE b.bucket_id = $1 AND b.user_id = $2
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

-- name: GetGeneralBucket :one
SELECT * FROM float.buckets WHERE user_id = $1 AND is_general = TRUE;
