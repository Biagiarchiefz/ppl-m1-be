-- name: CreateOrganizationApplication :one
INSERT INTO organization_applications (code, organization_name, admin_name, admin_email, password_hash, phone, description)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListOrganizationApplications :many
SELECT *
FROM organization_applications
WHERE (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountOrganizationApplications :one
SELECT count(*)
FROM organization_applications
WHERE (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text);

-- name: LockOrganizationApplication :one
SELECT *
FROM organization_applications
WHERE id = $1
FOR UPDATE;

-- name: ReviewOrganizationApplication :one
UPDATE organization_applications
SET status          = sqlc.arg(status)::text,
    status_reason   = sqlc.narg(status_reason),
    organization_id = sqlc.narg(organization_id),
    reviewed_by     = sqlc.narg(reviewed_by),
    reviewed_at     = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: GetLatestOrganizationApplication :one
SELECT *
FROM organization_applications
WHERE lower(admin_email) = lower($1)
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: OrganizationCodeExists :one
SELECT EXISTS (
  SELECT 1 FROM organizations WHERE lower(code) = lower($1)
);

-- name: GetOrganizationByCode :one
SELECT *
FROM organizations
WHERE lower(code) = lower($1);

-- name: DeleteOrganizationApplicationsByEmail :exec
DELETE FROM organization_applications
WHERE admin_email LIKE $1;

-- name: CountPermissions :one
SELECT count(*)
FROM permissions;
