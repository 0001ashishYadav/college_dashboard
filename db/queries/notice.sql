-- name: CreateNotice :one
INSERT INTO notices (
    institute_id,
    title,
    description,
    is_published,
    publish_date
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetNotice :one
SELECT *
FROM notices
WHERE id = $1 AND institute_id = $2
LIMIT 1;

-- name: GetNoticesByInstitute :many
SELECT *
FROM notices
WHERE institute_id = $1
ORDER BY created_at DESC;

-- name: UpdateNotice :one
UPDATE notices
SET
    title = $2,
    description = $3,
    is_published = $4,
    publish_date = $5
WHERE id = $1
RETURNING *;

-- name: DeleteNotice :exec
DELETE FROM notices
WHERE id = $1;

-- name: SearchNotices :many
SELECT *
FROM notices
WHERE institute_id = $1
AND title ILIKE '%' || $2 || '%'
ORDER BY created_at DESC;