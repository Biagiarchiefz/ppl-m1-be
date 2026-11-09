-- name: ListDivisions :many
SELECT *
FROM divisions
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(is_active)::boolean IS NULL OR is_active = sqlc.narg(is_active)::boolean)
ORDER BY code
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountDivisions :one
SELECT count(*)
FROM divisions
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(is_active)::boolean IS NULL OR is_active = sqlc.narg(is_active)::boolean);

-- name: GetDivisionInOrg :one
SELECT *
FROM divisions
WHERE id = $1 AND organization_id = $2;

-- name: LockDivisionInOrg :one
SELECT *
FROM divisions
WHERE id = $1 AND organization_id = $2
FOR UPDATE;

-- name: CreateDivision :one
INSERT INTO divisions (organization_id, code, name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateDivision :one
UPDATE divisions
SET name       = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeactivateDivision :one
UPDATE divisions
SET is_active  = FALSE,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CountActiveUsersWithDivision :one
SELECT count(*)
FROM users
WHERE division_id = $1 AND is_active;

-- name: ActivateDivision :exec
UPDATE divisions
SET is_active  = TRUE,
    updated_at = now()
WHERE id = $1;
