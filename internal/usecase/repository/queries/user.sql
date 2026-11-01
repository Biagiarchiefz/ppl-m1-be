-- name: GetRoleIDByName :one
SELECT id
FROM roles
WHERE LOWER(name) = LOWER(sqlc.arg(role_name)::text);

-- name: CreateUser :one
INSERT INTO users (email, password_hash, full_name, role_id)
VALUES ($1, $2, $3, $4)
RETURNING id, email, full_name;