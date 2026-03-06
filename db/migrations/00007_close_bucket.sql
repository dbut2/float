-- +goose Up
ALTER TABLE float.buckets ADD COLUMN status TEXT NOT NULL DEFAULT 'active';

-- +goose Down
ALTER TABLE float.buckets DROP COLUMN status;
