-- name: CreateUser :one
INSERT INTO users (organization_id, name, email, password_hash, role_id, division_id, unit_usaha_id, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: CreateSuperAdmin :one
INSERT INTO users (name, email, password_hash, is_super_admin)
VALUES ($1, $2, $3, TRUE)
RETURNING id;

-- name: GetUserDetailInOrg :one
SELECT *
FROM user_details
WHERE id = $1 AND organization_id = $2;

-- name: LockUserInOrg :one
SELECT *
FROM users
WHERE id = $1 AND organization_id = $2
FOR UPDATE;

-- name: UpdateUser :exec
UPDATE users
SET name          = $2,
    role_id       = $3,
    division_id   = $4,
    unit_usaha_id = $5,
    updated_at    = now()
WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users
SET is_active      = FALSE,
    deactivated_at = now(),
    updated_at     = now()
WHERE id = $1;

-- name: ActivateUser :exec
UPDATE users
SET is_active      = TRUE,
    deactivated_at = NULL,
    updated_at     = now()
WHERE id = $1;

-- name: ListUsers :many
SELECT *
FROM user_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(q)::text IS NULL
       OR name ILIKE '%' || sqlc.narg(q)::text || '%'
       OR email ILIKE '%' || sqlc.narg(q)::text || '%')
  AND (sqlc.narg(role)::text IS NULL OR role_code = sqlc.narg(role)::text)
  AND (sqlc.narg(division)::text IS NULL OR division_code = sqlc.narg(division)::text)
  AND (sqlc.narg(unit_usaha_id)::bigint IS NULL OR unit_usaha_id = sqlc.narg(unit_usaha_id)::bigint)
  AND (sqlc.narg(is_active)::boolean IS NULL OR is_active = sqlc.narg(is_active)::boolean)
ORDER BY
  CASE WHEN sqlc.arg(sort)::text = 'name' THEN name END ASC,
  CASE WHEN sqlc.arg(sort)::text = '-name' THEN name END DESC,
  CASE WHEN sqlc.arg(sort)::text = 'email' THEN email END ASC,
  CASE WHEN sqlc.arg(sort)::text = '-email' THEN email END DESC,
  CASE WHEN sqlc.arg(sort)::text = 'created_at' THEN created_at END ASC,
  CASE WHEN sqlc.arg(sort)::text = 'created_at' THEN id END ASC,
  created_at DESC,
  id DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountUsers :one
SELECT count(*)
FROM user_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(q)::text IS NULL
       OR name ILIKE '%' || sqlc.narg(q)::text || '%'
       OR email ILIKE '%' || sqlc.narg(q)::text || '%')
  AND (sqlc.narg(role)::text IS NULL OR role_code = sqlc.narg(role)::text)
  AND (sqlc.narg(division)::text IS NULL OR division_code = sqlc.narg(division)::text)
  AND (sqlc.narg(unit_usaha_id)::bigint IS NULL OR unit_usaha_id = sqlc.narg(unit_usaha_id)::bigint)
  AND (sqlc.narg(is_active)::boolean IS NULL OR is_active = sqlc.narg(is_active)::boolean);

-- name: GetActiveRoleIDByCode :one
SELECT id
FROM roles
WHERE organization_id = $1 AND code = $2 AND is_active;

-- name: GetActiveDivisionIDByCode :one
SELECT id
FROM divisions
WHERE organization_id = $1 AND code = $2 AND is_active;

-- name: DeleteUser :exec
-- Hapus permanen. Audit log tetap ada (user_id menjadi NULL, salinan
-- nama/email pelaku tetap tersimpan).
DELETE FROM users
WHERE id = $1;
