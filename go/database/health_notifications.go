package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const getBucketHealthNotification = `
SELECT notified_at FROM float.bucket_health_notifications
WHERE bucket_id = $1 AND user_id = $2
`

func (q *Queries) GetBucketHealthNotification(ctx context.Context, bucketID, userID uuid.UUID) (time.Time, error) {
	row := q.db.QueryRowContext(ctx, getBucketHealthNotification, bucketID, userID)
	var t time.Time
	err := row.Scan(&t)
	return t, err
}

const upsertBucketHealthNotification = `
INSERT INTO float.bucket_health_notifications (bucket_id, user_id, notified_at)
VALUES ($1, $2, now())
ON CONFLICT (bucket_id, user_id) DO UPDATE SET notified_at = now()
`

func (q *Queries) UpsertBucketHealthNotification(ctx context.Context, bucketID, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, upsertBucketHealthNotification, bucketID, userID)
	return err
}
