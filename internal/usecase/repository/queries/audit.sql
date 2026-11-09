-- name: InsertAuditLog :exec
-- Nama/email pelaku dan nama organisasi disalin saat kejadian supaya log
-- tetap terbaca setelah pengguna atau organisasinya dihapus permanen.
INSERT INTO audit_logs (organization_id, user_id, action, resource, resource_id, old_value, new_value,
                        ip_address, user_agent, actor_name, actor_email, organization_name)
VALUES (sqlc.narg(organization_id), sqlc.narg(user_id), sqlc.arg(action), sqlc.arg(resource),
        sqlc.narg(resource_id), sqlc.narg(old_value), sqlc.narg(new_value),
        sqlc.narg(ip_address), sqlc.narg(user_agent),
        (SELECT u.name FROM users u WHERE u.id = sqlc.narg(user_id)),
        (SELECT u.email FROM users u WHERE u.id = sqlc.narg(user_id)),
        (SELECT o.name FROM organizations o WHERE o.id = sqlc.narg(organization_id)));

-- name: ListAuditLogs :many
-- all_organizations = true (khusus super admin) menampilkan log semua
-- organisasi, termasuk organisasi yang sudah dihapus; organization_id lalu
-- menjadi filter opsional. Selain itu hanya log milik organization_id.
SELECT id, organization_id, organization_name, user_id, actor_name, actor_email,
       action, resource, resource_id, old_value, new_value, ip_address, created_at
FROM audit_logs
WHERE ((sqlc.arg(all_organizations)::boolean AND sqlc.narg(organization_id)::bigint IS NULL)
       OR organization_id = sqlc.narg(organization_id)::bigint)
  AND (sqlc.narg(from_time)::timestamptz IS NULL OR created_at >= sqlc.narg(from_time)::timestamptz)
  AND (sqlc.narg(to_time)::timestamptz IS NULL OR created_at < sqlc.narg(to_time)::timestamptz)
  AND (sqlc.narg(user_id)::bigint IS NULL OR user_id = sqlc.narg(user_id)::bigint)
  AND (sqlc.narg(action)::text IS NULL OR action = sqlc.narg(action)::text)
  AND (sqlc.narg(resource)::text IS NULL OR resource = sqlc.narg(resource)::text)
ORDER BY
  CASE WHEN sqlc.arg(sort)::text = 'created_at' THEN created_at END ASC,
  CASE WHEN sqlc.arg(sort)::text = 'created_at' THEN id END ASC,
  created_at DESC,
  id DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountAuditLogs :one
SELECT count(*)
FROM audit_logs
WHERE ((sqlc.arg(all_organizations)::boolean AND sqlc.narg(organization_id)::bigint IS NULL)
       OR organization_id = sqlc.narg(organization_id)::bigint)
  AND (sqlc.narg(from_time)::timestamptz IS NULL OR created_at >= sqlc.narg(from_time)::timestamptz)
  AND (sqlc.narg(to_time)::timestamptz IS NULL OR created_at < sqlc.narg(to_time)::timestamptz)
  AND (sqlc.narg(user_id)::bigint IS NULL OR user_id = sqlc.narg(user_id)::bigint)
  AND (sqlc.narg(action)::text IS NULL OR action = sqlc.narg(action)::text)
  AND (sqlc.narg(resource)::text IS NULL OR resource = sqlc.narg(resource)::text);
