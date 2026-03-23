-- name: UpsertUser :one
INSERT INTO float.users (email)
VALUES ($1)
ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
RETURNING *;

-- name: SeedUser :one
INSERT INTO float.users (user_id, email)
VALUES ($1, $2)
ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM float.users
WHERE user_id = $1;

-- name: GetUserByEmail :one
SELECT * FROM float.users
WHERE email = $1;

-- name: SetUserToken :exec
UPDATE float.users SET up_token = $2 WHERE user_id = $1;

-- name: GetUserToken :one
SELECT up_token FROM float.users WHERE user_id = $1;

-- name: SetUserWebhookSecret :exec
UPDATE float.users SET webhook_secret = $2 WHERE user_id = $1;

-- name: GetUserWebhookSecret :one
SELECT webhook_secret FROM float.users WHERE user_id = $1;
