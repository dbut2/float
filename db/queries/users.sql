-- name: UpsertUser :one
INSERT INTO float.users (email)
VALUES ($1)
ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM float.users
WHERE user_id = $1;

-- name: GetUserByEmail :one
SELECT * FROM float.users
WHERE email = $1;

-- name: UpsertUserToken :exec
INSERT INTO float.user_tokens (user_id, up_token)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET up_token = EXCLUDED.up_token;

-- name: GetUserToken :one
SELECT up_token FROM float.user_tokens
WHERE user_id = $1;

