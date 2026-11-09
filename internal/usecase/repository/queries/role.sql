-- name: ListRoles :many
SELECT *
FROM roles
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(is_active)::boolean IS NULL OR is_active = sqlc.narg(is_active)::boolean)
ORDER BY is_system DESC, code
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountRoles :one
SELECT count(*)
FROM roles
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(is_active)::boolean IS NULL OR is_active = sqlc.narg(is_active)::boolean);

-- name: ListOrganizationRolePermissions :many
SELECT rp.role_id, rp.permission_code
FROM role_permissions rp
JOIN roles r ON r.id = rp.role_id
WHERE r.organization_id = $1
ORDER BY rp.role_id, rp.permission_code;

-- name: GetRoleInOrg :one
SELECT *
FROM roles
WHERE id = $1 AND organization_id = $2;

-- name: LockRoleInOrg :one
SELECT *
FROM roles
WHERE id = $1 AND organization_id = $2
FOR UPDATE;

-- name: CreateRole :one
INSERT INTO roles (organization_id, code, name, description, is_system)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateRole :one
UPDATE roles
SET name        = $2,
    description = $3,
    updated_at  = now()
WHERE id = $1
RETURNING *;

-- name: DeactivateRole :one
UPDATE roles
SET is_active  = FALSE,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CountActiveUsersWithRole :one
SELECT count(*)
FROM users
WHERE role_id = $1 AND is_active;

-- name: DeleteRolePermissions :exec
DELETE FROM role_permissions
WHERE role_id = $1;

-- name: AddRolePermission :exec
INSERT INTO role_permissions (role_id, permission_code)
VALUES ($1, $2);

-- name: ListPermissions :many
SELECT *
FROM permissions
WHERE module <> 'platform'
ORDER BY module, code;

-- name: UpsertPermission :exec
INSERT INTO permissions (code, module, description)
VALUES ($1, $2, $3)
ON CONFLICT (code) DO UPDATE
SET module = EXCLUDED.module,
    description = EXCLUDED.description;

-- name: ActivateRole :exec
UPDATE roles
SET is_active  = TRUE,
    updated_at = now()
WHERE id = $1;
