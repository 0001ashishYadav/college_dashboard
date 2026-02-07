-- name: CreatePhoto :one
INSERT INTO photos (
    image_url,
    alt_text,
    uploaded_by,
    institute_id
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;


-- name: GetPhotoByID :one
SELECT *
FROM photos
WHERE id = $1
AND institute_id = $2
LIMIT 1;



-- name: GetPhotosByInstitute :many
SELECT *
FROM photos
WHERE institute_id = $1
ORDER BY created_at DESC;



-- name: UpdatePhoto :one
UPDATE photos
SET
    image_url = $3,
    alt_text = $4
WHERE id = $1
AND institute_id = $2
RETURNING *;



-- name: DeletePhoto :exec
DELETE FROM photos
WHERE id = $1
AND institute_id = $2;



-- name: GetPhotosByUser :many
SELECT *
FROM photos
WHERE uploaded_by = $1
AND institute_id = $2
ORDER BY created_at DESC;

