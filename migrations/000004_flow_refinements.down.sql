-- Mengembalikan ke bentuk 000003. Data yang tidak muat di bentuk lama dihapus:
-- invoice dari Order, Order batal, dan produk partner tanpa unit usaha.

DROP VIEW service_request_details;
DROP VIEW order_details;
DROP VIEW invoice_details;
DROP VIEW quotation_details;
DROP VIEW product_details;

DROP TABLE partner_customers;

DROP INDEX service_requests_user_idx;
ALTER TABLE service_requests
  DROP CONSTRAINT service_requests_status_check,
  DROP COLUMN status_changed_at;

ALTER TABLE quotations
  DROP COLUMN response_source,
  DROP COLUMN response_note;

DELETE FROM invoices WHERE order_id IS NOT NULL OR status = 'dibatalkan';
ALTER TABLE invoices
  DROP CONSTRAINT invoices_source_chk,
  DROP CONSTRAINT invoices_status_check,
  ADD CONSTRAINT invoices_status_check CHECK (status IN ('belum_lunas', 'lunas')),
  DROP COLUMN order_id,
  ALTER COLUMN quotation_id SET NOT NULL;

DELETE FROM orders WHERE status = 'dibatalkan';
ALTER TABLE orders
  DROP CONSTRAINT orders_status_check,
  ADD CONSTRAINT orders_status_check CHECK (status IN ('baru', 'diproses', 'selesai')),
  DROP CONSTRAINT orders_business_unit_fkey,
  DROP COLUMN business_unit_id;

DELETE FROM dashboard_metrics WHERE business_unit_id IS NULL;
DROP INDEX dashboard_metrics_period_uidx;
ALTER TABLE dashboard_metrics
  ALTER COLUMN business_unit_id SET NOT NULL,
  ADD UNIQUE (organization_id, business_unit_id, source_module, metric_key, period_date);

ALTER TABLE products DROP COLUMN stock;
DELETE FROM quotation_items WHERE business_unit_id IS NULL;
DELETE FROM order_items WHERE business_unit_id IS NULL;
DELETE FROM products WHERE business_unit_id IS NULL;
ALTER TABLE products ALTER COLUMN business_unit_id SET NOT NULL;
ALTER TABLE quotation_items ALTER COLUMN business_unit_id SET NOT NULL;
ALTER TABLE order_items ALTER COLUMN business_unit_id SET NOT NULL;

CREATE VIEW product_details AS
SELECT p.id, p.organization_id, p.business_unit_id, p.partner_id, p.type, p.name,
       p.category, p.description, p.price, p.is_active, p.created_at, p.updated_at,
       b.name AS business_unit_name, b.is_active AS business_unit_is_active,
       pa.business_name AS partner_name, pa.status AS partner_status
FROM products p
JOIN business_units b ON b.id = p.business_unit_id
LEFT JOIN partners pa ON pa.id = p.partner_id;

CREATE VIEW quotation_details AS
SELECT q.id, q.organization_id, q.partner_id, q.customer_id, q.status, q.notes,
       q.total, q.responded_at, q.created_at, c.name AS customer_name,
       pa.business_name AS partner_name, i.id AS invoice_id
FROM quotations q
JOIN customers c ON c.id = q.customer_id
LEFT JOIN partners pa ON pa.id = q.partner_id
LEFT JOIN invoices i ON i.quotation_id = q.id;

CREATE VIEW invoice_details AS
SELECT i.id, i.organization_id, i.quotation_id, i.partner_id, i.customer_id,
       i.total, i.status, i.paid_at, i.created_at, c.name AS customer_name,
       pa.business_name AS partner_name
FROM invoices i
JOIN customers c ON c.id = i.customer_id
LEFT JOIN partners pa ON pa.id = i.partner_id;

CREATE VIEW order_details AS
SELECT o.id, o.organization_id, o.partner_id, o.customer_id, o.status, o.total,
       o.status_changed_at, o.created_at, c.name AS customer_name,
       pa.business_name AS partner_name
FROM orders o
JOIN customers c ON c.id = o.customer_id
LEFT JOIN partners pa ON pa.id = o.partner_id;

CREATE VIEW service_request_details AS
SELECT s.id, s.organization_id, s.business_unit_id, s.user_id, s.contact_name,
       s.contact_email, s.contact_phone, s.subject, s.message, s.status,
       s.created_at, b.name AS business_unit_name
FROM service_requests s
JOIN business_units b ON b.id = s.business_unit_id;
