-- name: GetBucketHealthNotification :one
SELECT health_notified_at FROM float.buckets
WHERE bucket_id = $1 AND user_id = $2;

-- name: SetBucketHealthNotifiedAt :exec
UPDATE float.buckets SET health_notified_at = now()
WHERE bucket_id = $1 AND user_id = $2;
