-- name: ListProducts :many
SELECT *
FROM product_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(partner_id)::bigint IS NULL OR partner_id = sqlc.narg(partner_id)::bigint)
  AND (sqlc.narg(q)::text IS NULL OR name ILIKE '%' || sqlc.narg(q)::text || '%')
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountProducts :one
SELECT count(*)
FROM product_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(partner_id)::bigint IS NULL OR partner_id = sqlc.narg(partner_id)::bigint)
  AND (sqlc.narg(q)::text IS NULL OR name ILIKE '%' || sqlc.narg(q)::text || '%');

-- name: GetProduct :one
SELECT *
FROM product_details
WHERE id = $1 AND organization_id = $2;

-- name: LockProduct :one
SELECT *
FROM products
WHERE id = $1 AND organization_id = $2
FOR UPDATE;

-- name: CreateProduct :one
INSERT INTO products (organization_id, business_unit_id, partner_id, type, name, category, description, price, stock)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;

-- name: UpdateProduct :exec
UPDATE products
SET business_unit_id = $3,
    type             = $4,
    name             = $5,
    category         = $6,
    description      = $7,
    price            = $8,
    stock            = $9,
    updated_at       = now()
WHERE id = $1 AND organization_id = $2;

-- name: SetProductActive :exec
UPDATE products
SET is_active = $3, updated_at = now()
WHERE id = $1 AND organization_id = $2;

-- name: ListCustomers :many
SELECT *
FROM customers
WHERE customers.organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(q)::text IS NULL OR name ILIKE '%' || sqlc.narg(q)::text || '%' OR email ILIKE '%' || sqlc.narg(q)::text || '%')
  AND (sqlc.narg(partner_id)::bigint IS NULL OR EXISTS (
    SELECT 1 FROM partner_customers pc WHERE pc.customer_id = customers.id AND pc.partner_id = sqlc.narg(partner_id)::bigint))
ORDER BY name, id
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountCustomers :one
SELECT count(*)
FROM customers
WHERE customers.organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(q)::text IS NULL OR name ILIKE '%' || sqlc.narg(q)::text || '%' OR email ILIKE '%' || sqlc.narg(q)::text || '%')
  AND (sqlc.narg(partner_id)::bigint IS NULL OR EXISTS (
    SELECT 1 FROM partner_customers pc WHERE pc.customer_id = customers.id AND pc.partner_id = sqlc.narg(partner_id)::bigint));

-- name: GetCustomer :one
SELECT *
FROM customers
WHERE id = $1 AND organization_id = $2;

-- name: GetCustomerForUser :one
-- Mengutamakan pelanggan yang sudah tertaut ke akun; bila belum ada, cocokkan
-- dengan email akun.
SELECT c.*
FROM customers c
JOIN users u ON u.id = sqlc.arg(user_id) AND u.organization_id = c.organization_id
WHERE c.user_id = u.id OR lower(c.email) = lower(u.email)
ORDER BY (c.user_id IS NOT NULL) DESC
LIMIT 1;

-- name: CreateCustomerFromUser :one
INSERT INTO customers (organization_id, user_id, name, email)
SELECT u.organization_id, u.id, u.name, u.email
FROM users u
WHERE u.id = $1 AND u.organization_id IS NOT NULL
RETURNING *;

-- name: LinkCustomerUser :exec
UPDATE customers
SET user_id = $2, updated_at = now()
WHERE id = $1;

-- name: CreateCustomer :one
INSERT INTO customers (organization_id, name, email, phone, address)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateCustomer :one
UPDATE customers
SET name = $3, email = $4, phone = $5, address = $6, updated_at = now()
WHERE id = $1 AND organization_id = $2
RETURNING *;

-- name: ListQuotations :many
SELECT *
FROM quotation_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(partner_id)::bigint IS NULL OR partner_id = sqlc.narg(partner_id)::bigint)
  AND (sqlc.narg(customer_id)::bigint IS NULL OR customer_id = sqlc.narg(customer_id)::bigint)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountQuotations :one
SELECT count(*)
FROM quotation_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(partner_id)::bigint IS NULL OR partner_id = sqlc.narg(partner_id)::bigint)
  AND (sqlc.narg(customer_id)::bigint IS NULL OR customer_id = sqlc.narg(customer_id)::bigint)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text);

-- name: GetQuotation :one
SELECT *
FROM quotation_details
WHERE id = $1 AND organization_id = $2;

-- name: LockQuotation :one
SELECT *
FROM quotations
WHERE id = $1 AND organization_id = $2
FOR UPDATE;

-- name: CreateQuotation :one
INSERT INTO quotations (organization_id, partner_id, customer_id, notes, total, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: CreateQuotationItem :exec
INSERT INTO quotation_items (quotation_id, product_id, business_unit_id, product_name, price, qty, subtotal)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListQuotationItems :many
SELECT *
FROM quotation_items
WHERE quotation_id = $1
ORDER BY id;

-- name: UpdateQuotationStatus :exec
UPDATE quotations
SET status = $3, response_source = $4, response_note = $5, responded_at = now(), updated_at = now()
WHERE id = $1 AND organization_id = $2;

-- name: ListInvoices :many
SELECT *
FROM invoice_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(partner_id)::bigint IS NULL OR partner_id = sqlc.narg(partner_id)::bigint)
  AND (sqlc.narg(customer_id)::bigint IS NULL OR customer_id = sqlc.narg(customer_id)::bigint)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountInvoices :one
SELECT count(*)
FROM invoice_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(partner_id)::bigint IS NULL OR partner_id = sqlc.narg(partner_id)::bigint)
  AND (sqlc.narg(customer_id)::bigint IS NULL OR customer_id = sqlc.narg(customer_id)::bigint)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text);

-- name: GetInvoice :one
SELECT *
FROM invoice_details
WHERE id = $1 AND organization_id = $2;

-- name: LockInvoice :one
SELECT *
FROM invoices
WHERE id = $1 AND organization_id = $2
FOR UPDATE;

-- name: CreateInvoice :one
INSERT INTO invoices (organization_id, quotation_id, partner_id, customer_id, total, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: MarkInvoicePaid :exec
UPDATE invoices
SET status = 'lunas', paid_at = now(), updated_at = now()
WHERE id = $1;

-- name: CreatePayment :one
INSERT INTO payments (invoice_id, provider, reference, amount, status, message)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListInvoiceUnitRevenue :many
SELECT src.business_unit_id, sum(src.subtotal)::bigint AS revenue
FROM (
  SELECT qi.business_unit_id, qi.subtotal FROM quotation_items qi WHERE qi.quotation_id = sqlc.narg(quotation_id)::bigint
  UNION ALL
  SELECT oi.business_unit_id, oi.subtotal FROM order_items oi WHERE oi.order_id = sqlc.narg(order_id)::bigint
) src
GROUP BY src.business_unit_id;

-- name: CreateOrderInvoice :exec
INSERT INTO invoices (organization_id, order_id, partner_id, customer_id, total)
VALUES ($1, $2, $3, $4, $5);

-- name: LockInvoiceByOrder :one
SELECT *
FROM invoices
WHERE order_id = $1
FOR UPDATE;

-- name: CancelInvoice :exec
UPDATE invoices
SET status = 'dibatalkan', updated_at = now()
WHERE id = $1;

-- name: ListOrders :many
SELECT *
FROM order_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(business_unit_id)::bigint IS NULL OR business_unit_id = sqlc.narg(business_unit_id)::bigint)
  AND (sqlc.narg(partner_id)::bigint IS NULL OR partner_id = sqlc.narg(partner_id)::bigint)
  AND (sqlc.narg(customer_id)::bigint IS NULL OR customer_id = sqlc.narg(customer_id)::bigint)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CountOrders :one
SELECT count(*)
FROM order_details
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(business_unit_id)::bigint IS NULL OR business_unit_id = sqlc.narg(business_unit_id)::bigint)
  AND (sqlc.narg(partner_id)::bigint IS NULL OR partner_id = sqlc.narg(partner_id)::bigint)
  AND (sqlc.narg(customer_id)::bigint IS NULL OR customer_id = sqlc.narg(customer_id)::bigint)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text);

-- name: GetOrder :one
SELECT *
FROM order_details
WHERE id = $1 AND organization_id = $2;

-- name: LockOrder :one
SELECT *
FROM orders
WHERE id = $1 AND organization_id = $2
FOR UPDATE;

-- name: CreateOrder :one
INSERT INTO orders (organization_id, partner_id, business_unit_id, customer_id, total)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: CreateOrderItem :exec
INSERT INTO order_items (order_id, product_id, business_unit_id, product_name, price, qty, subtotal)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListOrderItems :many
SELECT *
FROM order_items
WHERE order_id = $1
ORDER BY id;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $3, status_changed_at = now()
WHERE id = $1 AND organization_id = $2;

-- name: ReserveStock :execrows
-- Baris tidak berubah bila stok kurang; jasa (stok NULL) selalu lolos.
UPDATE products
SET stock = CASE WHEN stock IS NULL THEN NULL ELSE stock - sqlc.arg(qty)::int END, updated_at = now()
WHERE id = sqlc.arg(id) AND (stock IS NULL OR stock >= sqlc.arg(qty)::int);

-- name: RestoreStock :exec
UPDATE products
SET stock = stock + sqlc.arg(qty)::int, updated_at = now()
WHERE id = sqlc.arg(id) AND stock IS NOT NULL;

-- name: GetCustomerByEmail :one
SELECT *
FROM customers
WHERE organization_id = $1 AND lower(email) = lower($2);

-- name: LinkPartnerCustomer :exec
INSERT INTO partner_customers (organization_id, partner_id, customer_id)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: PartnerHasCustomer :one
SELECT EXISTS (
  SELECT 1 FROM partner_customers WHERE partner_id = $1 AND customer_id = $2
);
