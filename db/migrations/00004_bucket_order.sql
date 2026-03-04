-- +goose Up
ALTER TABLE float.buckets ADD COLUMN display_order INTEGER;

-- +goose Down
ALTER TABLE float.buckets DROP COLUMN display_order;
