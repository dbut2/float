-- name: GetBucketHealthNotification :one
SELECT notified_at FROM float.bucket_health_notifications
WHERE bucket_id = $1 AND user_id = $2;

-- name: UpsertBucketHealthNotification :exec
INSERT INTO float.bucket_health_notifications (bucket_id, user_id, notified_at)
VALUES ($1, $2, now())
ON CONFLICT (bucket_id, user_id) DO UPDATE SET notified_at = now();
