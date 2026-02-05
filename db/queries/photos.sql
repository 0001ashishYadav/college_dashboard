-- name: CreatePhoto :one
INSERT INTO photos (
    image_url,
    alt_text,
    uploaded_by
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetPhotoByID :one
SELECT *
FROM photos
WHERE id = $1
LIMIT 1;


-- name: GetAllPhotos :many
SELECT *
FROM photos
ORDER BY created_at DESC;


-- name: UpdatePhoto :one
UPDATE photos
SET
    image_url = $2,
    alt_text = $3
WHERE id = $1
RETURNING *;


-- name: DeletePhoto :exec
DELETE FROM photos
WHERE id = $1;


-- name: GetPhotosByUser :many
SELECT *
FROM photos
WHERE uploaded_by = $1
ORDER BY created_at DESC;
