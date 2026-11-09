-- name: ListBusinessUnits :many
SELECT *
FROM business_unit_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(is_active)::boolean IS NULL OR is_active = sqlc.narg(is_active)::boolean)
  AND (sqlc.narg(q)::text IS NULL OR name ILIKE '%' || sqlc.narg(q)::text || '%' OR code ILIKE '%' || sqlc.narg(q)::text || '%')
ORDER BY name, id
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountBusinessUnits :one
SELECT count(*)
FROM business_unit_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(is_active)::boolean IS NULL OR is_active = sqlc.narg(is_active)::boolean)
  AND (sqlc.narg(q)::text IS NULL OR name ILIKE '%' || sqlc.narg(q)::text || '%' OR code ILIKE '%' || sqlc.narg(q)::text || '%');

-- name: GetBusinessUnit :one
SELECT *
FROM business_unit_details
WHERE id = $1 AND organization_id = $2;

-- name: LockBusinessUnit :one
SELECT *
FROM business_units
WHERE id = $1 AND organization_id = $2
FOR UPDATE;

-- name: ActiveDivisionExists :one
SELECT EXISTS (
  SELECT 1 FROM divisions WHERE id = $1 AND organization_id = $2 AND is_active
);

-- name: CreateBusinessUnit :one
INSERT INTO business_units (organization_id, division_id, code, name, category, description, address, phone, email)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;

-- name: UpdateBusinessUnit :exec
UPDATE business_units
SET division_id = $3,
    name        = $4,
    category    = $5,
    description = $6,
    address     = $7,
    phone       = $8,
    email       = $9,
    updated_at  = now()
WHERE id = $1 AND organization_id = $2;

-- name: SetBusinessUnitActive :exec
UPDATE business_units
SET is_active = $3, updated_at = now()
WHERE id = $1 AND organization_id = $2;

-- name: DeleteOrganizationBusinessUnits :exec
-- Produk, metrik, dan permintaan layanan ikut terhapus lewat ON DELETE CASCADE.
DELETE FROM business_units
WHERE organization_id = $1;

-- name: UpsertDashboardMetric :exec
INSERT INTO dashboard_metrics (organization_id, business_unit_id, source_module, metric_key, label, unit, value, period_date)
VALUES (
  sqlc.arg(organization_id), sqlc.narg(business_unit_id), sqlc.arg(source_module), sqlc.arg(metric_key),
  sqlc.arg(label), sqlc.arg(unit), sqlc.arg(value),
  COALESCE(sqlc.narg(period_date)::date, (now() AT TIME ZONE 'Asia/Jakarta')::date)
)
ON CONFLICT (organization_id, COALESCE(business_unit_id, 0), source_module, metric_key, period_date)
DO UPDATE SET value      = dashboard_metrics.value + EXCLUDED.value,
              label      = EXCLUDED.label,
              unit       = EXCLUDED.unit,
              updated_at = now();

-- name: ReportDashboardMetrics :many
SELECT m.business_unit_id, COALESCE(b.name, 'Penjualan partner')::text AS business_unit_name, m.source_module, m.metric_key,
       m.label, m.unit, sum(m.value)::bigint AS total
FROM dashboard_metrics m
LEFT JOIN business_units b ON b.id = m.business_unit_id
WHERE m.organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(business_unit_id)::bigint IS NULL OR m.business_unit_id = sqlc.narg(business_unit_id)::bigint)
  AND (sqlc.narg(date_from)::date IS NULL OR m.period_date >= sqlc.narg(date_from)::date)
  AND (sqlc.narg(date_to)::date IS NULL OR m.period_date <= sqlc.narg(date_to)::date)
GROUP BY m.business_unit_id, b.name, m.source_module, m.metric_key, m.label, m.unit
ORDER BY b.name, m.source_module, m.metric_key;
