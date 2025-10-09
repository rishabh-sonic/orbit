-- name: GetConfigValue :one
SELECT value FROM site_config WHERE key = $1 LIMIT 1;

-- name: UpsertConfigValue :exec
INSERT INTO site_config (key, value, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();

-- name: ListConfig :many
SELECT * FROM site_config ORDER BY key;
