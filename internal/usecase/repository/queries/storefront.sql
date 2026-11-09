-- name: ListCatalogBusinessUnits :many
SELECT *
FROM business_unit_details
WHERE organization_id = $1 AND is_active
ORDER BY name, id;

-- name: ListCatalogProducts :many
SELECT *
FROM product_details
WHERE organization_id = sqlc.arg(organization_id)
  AND is_active AND (business_unit_id IS NULL OR business_unit_is_active)
  AND (partner_id IS NULL OR partner_status = 'terverifikasi')
  AND (sqlc.narg(q)::text IS NULL OR name ILIKE '%' || sqlc.narg(q)::text || '%' OR description ILIKE '%' || sqlc.narg(q)::text || '%')
  AND (sqlc.narg(category)::text IS NULL OR category = sqlc.narg(category)::text)
  AND (sqlc.narg(partner_id)::bigint IS NULL OR partner_id = sqlc.narg(partner_id)::bigint)
  AND (sqlc.narg(business_unit_id)::bigint IS NULL OR business_unit_id = sqlc.narg(business_unit_id)::bigint)
ORDER BY name, id
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountCatalogProducts :one
SELECT count(*)
FROM product_details
WHERE organization_id = sqlc.arg(organization_id)
  AND is_active AND (business_unit_id IS NULL OR business_unit_is_active)
  AND (partner_id IS NULL OR partner_status = 'terverifikasi')
  AND (sqlc.narg(q)::text IS NULL OR name ILIKE '%' || sqlc.narg(q)::text || '%' OR description ILIKE '%' || sqlc.narg(q)::text || '%')
  AND (sqlc.narg(category)::text IS NULL OR category = sqlc.narg(category)::text)
  AND (sqlc.narg(partner_id)::bigint IS NULL OR partner_id = sqlc.narg(partner_id)::bigint)
  AND (sqlc.narg(business_unit_id)::bigint IS NULL OR business_unit_id = sqlc.narg(business_unit_id)::bigint);

-- name: ListCatalogCategories :many
SELECT DISTINCT category
FROM product_details
WHERE organization_id = $1
  AND is_active AND (business_unit_id IS NULL OR business_unit_is_active)
  AND (partner_id IS NULL OR partner_status = 'terverifikasi')
ORDER BY category;

-- name: GetCatalogProduct :one
SELECT *
FROM product_details
WHERE id = $1 AND organization_id = $2
  AND is_active AND (business_unit_id IS NULL OR business_unit_is_active)
  AND (partner_id IS NULL OR partner_status = 'terverifikasi');

-- name: AddCartItem :exec
INSERT INTO cart_items (user_id, product_id, qty)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, product_id)
DO UPDATE SET qty = cart_items.qty + EXCLUDED.qty, updated_at = now();

-- name: SetCartItemQty :execrows
UPDATE cart_items
SET qty = $3, updated_at = now()
WHERE user_id = $1 AND product_id = $2;

-- name: DeleteCartItem :execrows
DELETE FROM cart_items
WHERE user_id = $1 AND product_id = $2;

-- name: ListCartItems :many
SELECT ci.product_id, ci.qty, p.name, p.price, p.business_unit_id, p.business_unit_name, p.partner_id,
       p.stock, p.partner_name,
       (p.is_active AND (p.business_unit_id IS NULL OR p.business_unit_is_active) AND (p.partner_id IS NULL OR p.partner_status = 'terverifikasi') AND (p.stock IS NULL OR p.stock > 0))::boolean AS available
FROM cart_items ci
JOIN product_details p ON p.id = ci.product_id
WHERE ci.user_id = $1
ORDER BY ci.created_at, ci.product_id;

-- name: LockCartItems :many
-- Mengunci keranjang agar dua checkout bersamaan tidak membuat Order ganda.
SELECT ci.product_id, ci.qty, p.name, p.price, p.business_unit_id, p.business_unit_name, p.partner_id,
       p.stock, p.partner_name,
       (p.is_active AND (p.business_unit_id IS NULL OR p.business_unit_is_active) AND (p.partner_id IS NULL OR p.partner_status = 'terverifikasi') AND (p.stock IS NULL OR p.stock > 0))::boolean AS available
FROM cart_items ci
JOIN product_details p ON p.id = ci.product_id
WHERE ci.user_id = $1
ORDER BY ci.product_id
FOR UPDATE OF ci;

-- name: ClearCart :exec
DELETE FROM cart_items
WHERE user_id = $1;

-- name: CreateServiceRequest :one
INSERT INTO service_requests (organization_id, business_unit_id, user_id, contact_name, contact_email, contact_phone, subject, message)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListServiceRequests :many
SELECT *
FROM service_request_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(business_unit_id)::bigint IS NULL OR business_unit_id = sqlc.narg(business_unit_id)::bigint)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountServiceRequests :one
SELECT count(*)
FROM service_request_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(business_unit_id)::bigint IS NULL OR business_unit_id = sqlc.narg(business_unit_id)::bigint);

-- name: LockServiceRequest :one
SELECT *
FROM service_requests
WHERE id = $1 AND organization_id = $2
FOR UPDATE;

-- name: UpdateServiceRequestStatus :exec
UPDATE service_requests
SET status = $3, status_changed_at = now()
WHERE id = $1 AND organization_id = $2;

-- name: ListMyServiceRequests :many
SELECT *
FROM service_request_details
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountMyServiceRequests :one
SELECT count(*)
FROM service_request_details
WHERE user_id = $1;

-- name: SetCustomerPhoneIfEmpty :exec
UPDATE customers
SET phone = COALESCE(phone, $2), updated_at = now()
WHERE id = $1;
