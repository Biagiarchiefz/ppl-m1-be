-- name: ListOrganizations :many
SELECT *
FROM organizations
WHERE (sqlc.narg(q)::text IS NULL
       OR name ILIKE '%' || sqlc.narg(q)::text || '%'
       OR code ILIKE '%' || sqlc.narg(q)::text || '%')
  AND (sqlc.narg(is_active)::boolean IS NULL OR is_active = sqlc.narg(is_active)::boolean)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountOrganizations :one
SELECT count(*)
FROM organizations
WHERE (sqlc.narg(q)::text IS NULL
       OR name ILIKE '%' || sqlc.narg(q)::text || '%'
       OR code ILIKE '%' || sqlc.narg(q)::text || '%')
  AND (sqlc.narg(is_active)::boolean IS NULL OR is_active = sqlc.narg(is_active)::boolean);

-- name: GetOrganization :one
SELECT *
FROM organizations
WHERE id = $1;

-- name: LockOrganization :one
SELECT *
FROM organizations
WHERE id = $1
FOR UPDATE;

-- name: CreateOrganization :one
INSERT INTO organizations (code, name)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateOrganization :one
UPDATE organizations
SET name       = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeactivateOrganization :one
UPDATE organizations
SET is_active      = FALSE,
    deactivated_at = now(),
    updated_at     = now()
WHERE id = $1
RETURNING *;

-- name: CountOrganizationUsers :one
SELECT count(*)
FROM users
WHERE organization_id = $1;

-- name: DeleteOrganizationUsers :exec
DELETE FROM users
WHERE organization_id = $1;

-- name: DeleteOrganizationRoles :exec
-- role_permissions ikut terhapus lewat ON DELETE CASCADE.
DELETE FROM roles
WHERE organization_id = $1;

-- name: DeleteOrganizationDivisions :exec
DELETE FROM divisions
WHERE organization_id = $1;

-- name: DeleteOrganization :exec
DELETE FROM organizations
WHERE id = $1;

-- name: ActivateOrganization :exec
UPDATE organizations
SET is_active      = TRUE,
    deactivated_at = NULL,
    updated_at     = now()
WHERE id = $1;
