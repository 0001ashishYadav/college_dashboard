-- name: CreateCarousel :one
INSERT INTO carousels (
    institute_id,
    title,
    is_active
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetCarousel :one
SELECT *
FROM carousels
WHERE id = $1
AND institute_id = $2
LIMIT 1;

-- name: GetCarouselsByInstitute :many
SELECT *
FROM carousels
WHERE institute_id = $1
ORDER BY created_at DESC;

-- name: UpdateCarousel :one
UPDATE carousels
SET
    title = $3,
    is_active = $4
WHERE id = $1
AND institute_id = $2
RETURNING *;

-- name: DeleteCarousel :exec
DELETE FROM carousels
WHERE id = $1
AND institute_id = $2;
