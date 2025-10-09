-- name: UpsertImage :one
INSERT INTO images (url)
VALUES ($1)
ON CONFLICT (url) DO UPDATE SET ref_count = images.ref_count + 1
RETURNING *;

-- name: DecrementImageRefCount :exec
UPDATE images SET ref_count = ref_count - 1 WHERE url = $1;

-- name: GetImageByURL :one
SELECT * FROM images WHERE url = $1 LIMIT 1;
