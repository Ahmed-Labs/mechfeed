-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users 
WHERE id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
  id, username, webhook_url
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetAlerts :many
SELECT * FROM user_alerts;

-- name: GetUserAlerts :many
SELECT * FROM user_alerts
WHERE id = $1;

-- name: CreateAlert :exec
INSERT INTO user_alerts (
  id, keyword
) VALUES (
  $1, $2
);

-- name: DeleteAlert :exec
DELETE FROM user_alerts
WHERE alert_id = $1;
