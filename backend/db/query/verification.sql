-- name: CreateVerificationCode :one
INSERT INTO verification_codes (email, code, type, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetVerificationCode :one
SELECT * FROM verification_codes
WHERE email = $1 AND type = $2 AND used = FALSE AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkVerificationCodeUsed :exec
UPDATE verification_codes SET used = TRUE WHERE id = $1;

-- name: DeleteVerificationCodes :exec
DELETE FROM verification_codes WHERE email = $1 AND type = $2;
