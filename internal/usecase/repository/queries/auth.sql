-- name: GetUserDetailByEmail :one
SELECT *
FROM user_details
WHERE lower(email) = lower(sqlc.arg(email)::text);

-- name: GetUserDetailByID :one
SELECT *
FROM user_details
WHERE id = $1;

-- name: ListRolePermissionCodes :many
SELECT permission_code
FROM role_permissions
WHERE role_id = $1
ORDER BY permission_code;

-- name: ListOrganizationPermissionCodes :many
SELECT code
FROM permissions
WHERE module <> 'platform'
ORDER BY code;

-- name: LockUserByID :one
SELECT *
FROM users
WHERE id = $1
FOR UPDATE;

-- name: UpdatePassword :one
UPDATE users
SET password_hash = $2,
    updated_at    = now()
WHERE id = $1
RETURNING updated_at;
