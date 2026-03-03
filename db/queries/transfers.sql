-- name: CreateTransfer :one
INSERT INTO float.bucket_transfers (from_bucket_id, to_bucket_id, amount_cents, note)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListTransfers :many
SELECT
    t.transfer_id,
    t.from_bucket_id,
    fb.name AS from_bucket_name,
    t.to_bucket_id,
    tb.name AS to_bucket_name,
    t.amount_cents,
    t.note,
    t.created_at
FROM float.bucket_transfers t
JOIN float.buckets fb ON t.from_bucket_id = fb.bucket_id
JOIN float.buckets tb ON t.to_bucket_id = tb.bucket_id
WHERE fb.user_id = $1
ORDER BY t.created_at DESC;

-- name: DeleteTransfer :execrows
DELETE FROM float.bucket_transfers t
USING float.buckets b
WHERE t.transfer_id = $1 AND t.from_bucket_id = b.bucket_id AND b.user_id = $2;
