-- name: InsertTrickle :one
INSERT INTO float.bucket_trickles (from_bucket_id, to_bucket_id, amount_cents, description, period, start_date, end_date)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetActiveTrickleByToBucketID :one
SELECT t.*, tb.name AS to_bucket_name, fb.name AS from_bucket_name
FROM float.bucket_trickles t
JOIN float.buckets tb ON tb.bucket_id = t.to_bucket_id
JOIN float.buckets fb ON fb.bucket_id = t.from_bucket_id
WHERE t.to_bucket_id = $1 AND tb.user_id = $2
  AND (t.end_date IS NULL OR t.end_date >= CURRENT_DATE)
ORDER BY t.created_at DESC
LIMIT 1;

-- name: ListTrickles :many
SELECT t.*, tb.name AS to_bucket_name, fb.name AS from_bucket_name
FROM float.bucket_trickles t
JOIN float.buckets tb ON tb.bucket_id = t.to_bucket_id
JOIN float.buckets fb ON fb.bucket_id = t.from_bucket_id
WHERE tb.user_id = $1
  AND (t.end_date IS NULL OR t.end_date >= CURRENT_DATE)
ORDER BY t.created_at DESC;

-- name: GetTricklesByBucketID :many
SELECT t.*, tb.name AS to_bucket_name, fb.name AS from_bucket_name
FROM float.bucket_trickles t
JOIN float.buckets tb ON tb.bucket_id = t.to_bucket_id
JOIN float.buckets fb ON fb.bucket_id = t.from_bucket_id
WHERE t.to_bucket_id = $1 OR t.from_bucket_id = $1;

-- name: SetTrickleEndDate :exec
UPDATE float.bucket_trickles t
SET end_date = $2
FROM float.buckets b
WHERE t.trickle_id = $1 AND t.to_bucket_id = b.bucket_id AND b.user_id = $3;

-- name: DeleteTrickle :execrows
DELETE FROM float.bucket_trickles t
USING float.buckets b
WHERE t.trickle_id = $1 AND t.to_bucket_id = b.bucket_id AND b.user_id = $2;
