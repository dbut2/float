-- +goose Up
DROP TABLE float.bucket_health_notifications;
ALTER TABLE float.buckets ADD COLUMN health_notified_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE float.buckets DROP COLUMN health_notified_at;
CREATE TABLE float.bucket_health_notifications (
  bucket_id   UUID NOT NULL,
  user_id     UUID NOT NULL,
  notified_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (bucket_id, user_id)
);
