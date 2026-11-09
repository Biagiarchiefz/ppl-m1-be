-- name: GetActiveOrganizationByCode :one
SELECT *
FROM organizations
WHERE lower(code) = lower($1) AND is_active;

-- name: CreatePartner :one
INSERT INTO partners (organization_id, user_id, business_name, business_type, contact_name, email, phone, address, description)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListPartners :many
SELECT *
FROM partners
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
  AND (sqlc.narg(q)::text IS NULL OR business_name ILIKE '%' || sqlc.narg(q)::text || '%' OR email ILIKE '%' || sqlc.narg(q)::text || '%')
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountPartners :one
SELECT count(*)
FROM partners
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
  AND (sqlc.narg(q)::text IS NULL OR business_name ILIKE '%' || sqlc.narg(q)::text || '%' OR email ILIKE '%' || sqlc.narg(q)::text || '%');

-- name: GetPartner :one
SELECT *
FROM partners
WHERE id = $1 AND organization_id = $2;

-- name: GetPartnerByUser :one
SELECT *
FROM partners
WHERE user_id = $1;

-- name: LockPartner :one
SELECT *
FROM partners
WHERE id = $1 AND organization_id = $2
FOR UPDATE;

-- name: UpdatePartnerStatus :one
UPDATE partners
SET status        = sqlc.arg(status)::text,
    status_reason = sqlc.narg(status_reason),
    verified_at   = CASE WHEN sqlc.arg(status)::text = 'terverifikasi' THEN COALESCE(verified_at, now()) ELSE verified_at END,
    verified_by   = CASE WHEN sqlc.arg(status)::text = 'terverifikasi' THEN COALESCE(verified_by, sqlc.narg(actor_id)::bigint) ELSE verified_by END,
    updated_at    = now()
WHERE id = sqlc.arg(id) AND organization_id = sqlc.arg(organization_id)
RETURNING *;

-- name: LockPartnerByUser :one
SELECT *
FROM partners
WHERE user_id = $1
FOR UPDATE;

-- name: ResubmitPartner :one
UPDATE partners
SET business_name = $2,
    business_type = $3,
    contact_name  = $4,
    phone         = $5,
    address       = $6,
    description   = $7,
    status        = 'pending',
    status_reason = NULL,
    updated_at    = now()
WHERE id = $1
RETURNING *;

-- name: ListVerifiedPartners :many
SELECT *
FROM partners
WHERE organization_id = $1 AND status = 'terverifikasi'
ORDER BY business_name, id;

-- name: ListAllOrganizationIDs :many
SELECT id
FROM organizations
ORDER BY id;

-- name: EnsureRole :one
INSERT INTO roles (organization_id, code, name)
VALUES ($1, $2, $3)
ON CONFLICT (organization_id, code) DO UPDATE SET code = roles.code
RETURNING id;

-- name: GrantDefaultPermissions :exec
-- Hanya mengisi peran yang belum punya izin, supaya izin yang sudah diatur
-- administrator tidak ditimpa saat seeder dijalankan ulang.
INSERT INTO role_permissions (role_id, permission_code)
SELECT sqlc.arg(role_id)::bigint, p.code
FROM permissions p
WHERE p.code = ANY(string_to_array(sqlc.arg(codes)::text, ','))
  AND NOT EXISTS (SELECT 1 FROM role_permissions rp WHERE rp.role_id = sqlc.arg(role_id)::bigint);
