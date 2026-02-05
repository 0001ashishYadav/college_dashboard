-- name: CreateCarouselPhoto :one
INSERT INTO carousel_photos (
    carousel_id,
    photo_id,
    display_text,
    display_order
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetCarouselPhotoByID :one
SELECT *
FROM carousel_photos
WHERE id = $1
LIMIT 1;


-- name: GetCarouselPhotosByCarouselID :many
SELECT
    cp.id,
    cp.carousel_id,
    cp.photo_id,
    cp.display_text,
    cp.display_order,
    cp.created_at,
    p.image_url,
    p.alt_text
FROM carousel_photos cp
JOIN photos p ON p.id = cp.photo_id
WHERE cp.carousel_id = $1
ORDER BY cp.display_order ASC;


-- name: UpdateCarouselPhoto :one
UPDATE carousel_photos
SET
    display_text = $2,
    display_order = $3
WHERE id = $1
RETURNING *;


-- name: ReorderCarouselPhoto :exec
UPDATE carousel_photos
SET display_order = $2
WHERE id = $1;


-- name: DeleteCarouselPhoto :exec
DELETE FROM carousel_photos
WHERE id = $1;
