-- name: CreateInstitute :one
INSERT INTO institutes (
    name,
    code,
    email,
    phone,
    address,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetInstituteByID :one
SELECT *
FROM institutes
WHERE id = $1
AND is_active = true
LIMIT 1;


-- name: GetInstituteByCode :one
SELECT *
FROM institutes
WHERE code = $1
AND is_active = true
LIMIT 1;


-- name: GetAllInstitutes :many
SELECT *
FROM institutes
WHERE is_active = true
ORDER BY created_at DESC;


-- name: UpdateInstitute :one
UPDATE institutes
SET
    name = $2,
    code = $3,
    email = $4,
    phone = $5,
    address = $6,
    is_active = $7
WHERE id = $1
RETURNING *;


-- name: DisableInstitute :exec
UPDATE institutes
SET is_active = false
WHERE id = $1;


-- name: DeleteInstitute :exec
DELETE FROM institutes
WHERE id = $1;
