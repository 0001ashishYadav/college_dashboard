-- name: Login :one
SELECT 
name,
role,
email
FROM admin
WHERE id = $1 OR email = $2
LIMIT 1;