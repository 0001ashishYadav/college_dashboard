-- name: CreateCarousel :one
INSERT INTO carousels (
    institute_id,
    title,
    is_active
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetCarouselWithPhotos :many
SELECT
    c.id              AS carousel_id,
    c.institute_id,
    c.title,
    c.is_active,
    c.created_at,

    cp.id             AS carousel_photo_id,
    cp.display_text,
    cp.display_order,

    p.id              AS photo_id,
    p.image_url,
    p.alt_text
FROM carousels c
LEFT JOIN carousel_photos cp ON cp.carousel_id = c.id
LEFT JOIN photos p ON p.id = cp.photo_id
WHERE c.id = $1
  AND c.institute_id = $2
ORDER BY cp.display_order ASC;


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
