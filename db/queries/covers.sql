-- name: CreateCover :one
INSERT INTO float.bucket_transfers (from_bucket_id, to_bucket_id, amount_cents, note, covers_transaction_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListCoversByTransaction :many
SELECT transfer_id, amount_cents, note, created_at, covers_transaction_id
FROM float.bucket_transfers
WHERE covers_transaction_id = $1;

-- name: ListCoversByBucket :many
SELECT t.transfer_id, t.amount_cents, t.note, t.created_at, t.covers_transaction_id,
       tx.description AS transaction_description
FROM float.bucket_transfers t
JOIN float.up_transactions tx ON t.covers_transaction_id = tx.transaction_id
WHERE tx.bucket_id = $1;

-- name: GetTransactionOwner :one
SELECT tx.transaction_id, tx.bucket_id, tx.amount_cents, b.user_id
FROM float.up_transactions tx
JOIN float.buckets b ON tx.bucket_id = b.bucket_id
WHERE tx.transaction_id = $1 AND b.user_id = $2;

-- name: DeleteCover :execrows
DELETE FROM float.bucket_transfers t
USING float.up_transactions tx
JOIN float.buckets b ON tx.bucket_id = b.bucket_id
WHERE t.covers_transaction_id = tx.transaction_id
  AND t.transfer_id = $1
  AND b.user_id = $2;
