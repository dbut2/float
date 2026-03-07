-- name: InsertClassificationLog :exec
INSERT INTO float.classification_log (transaction_id, chosen_bucket_id, confidence, reasoning, model)
VALUES ($1, $2, $3, $4, $5);
