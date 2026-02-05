-- name: LoginUser :one
SELECT *
FROM users
WHERE (id = $1 OR email = $2)
AND is_active = true
LIMIT 1;


-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
AND institute_id = $2
LIMIT 1;


-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
AND is_active = true
LIMIT 1;


-- name: GetUsersByInstitute :many
SELECT *
FROM users
WHERE institute_id = $1
ORDER BY created_at DESC;


-- name: CreateUser :one
INSERT INTO users (
    institute_id,
    name,
    email,
    password,
    role,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;


-- name: UpdateUser :one
UPDATE users
SET
    name = $2,
    email = $3,
    role = $4,
    is_active = $5
WHERE id = $1
RETURNING *;


-- name: UpdateUserPassword :exec
UPDATE users
SET password = $2
WHERE id = $1;


-- name: DisableUser :exec
UPDATE users
SET is_active = false
WHERE id = $1;


-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
