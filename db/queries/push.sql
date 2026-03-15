-- name: RegisterFCMToken :exec
WITH deleted AS (
    DELETE FROM float.fcm_tokens WHERE user_id = $1
)
INSERT INTO float.fcm_tokens (user_id, fcm_token) VALUES ($1, $2)
ON CONFLICT (fcm_token) DO UPDATE SET user_id = EXCLUDED.user_id;

-- name: UnregisterFCMToken :exec
DELETE FROM float.fcm_tokens
WHERE user_id = $1 AND fcm_token = $2;

-- name: GetUserFCMTokens :many
SELECT fcm_token FROM float.fcm_tokens
WHERE user_id = $1;
