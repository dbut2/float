-- +goose Up
CREATE TABLE IF NOT EXISTS float.bucket_health_notifications (
  bucket_id UUID NOT NULL,
  user_id   UUID NOT NULL,
  notified_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (bucket_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS float.bucket_health_notifications;
