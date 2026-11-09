-- Penyempurnaan alur M2-M5: invoice dari Order, unit usaha opsional untuk
-- produk partner, stok, pembatalan Order, persetujuan quotation offline,
-- status permintaan layanan, dan pembatasan pelanggan per partner.

DROP VIEW service_request_details;
DROP VIEW order_details;
DROP VIEW invoice_details;
DROP VIEW quotation_details;
DROP VIEW product_details;

-- Produk partner boleh berdiri sendiri tanpa unit usaha.
ALTER TABLE products ALTER COLUMN business_unit_id DROP NOT NULL;
ALTER TABLE quotation_items ALTER COLUMN business_unit_id DROP NOT NULL;
ALTER TABLE order_items ALTER COLUMN business_unit_id DROP NOT NULL;

-- Stok hanya untuk barang; NULL berarti tidak dibatasi (jasa).
ALTER TABLE products ADD COLUMN stock INTEGER CHECK (stock >= 0);

-- Dashboard menerima data tanpa unit usaha (penjualan partner).
ALTER TABLE dashboard_metrics ALTER COLUMN business_unit_id DROP NOT NULL;
DO $$
DECLARE
  unique_name text;
BEGIN
  SELECT conname INTO unique_name
  FROM pg_constraint
  WHERE conrelid = 'dashboard_metrics'::regclass AND contype = 'u';
  EXECUTE format('ALTER TABLE dashboard_metrics DROP CONSTRAINT %I', unique_name);
END $$;
CREATE UNIQUE INDEX dashboard_metrics_period_uidx ON dashboard_metrics
  (organization_id, COALESCE(business_unit_id, 0), source_module, metric_key, period_date);

-- Order organisasi dipegang unit usahanya; Order partner dipegang partnernya.
ALTER TABLE orders
  ADD COLUMN business_unit_id BIGINT,
  ADD CONSTRAINT orders_business_unit_fkey FOREIGN KEY (organization_id, business_unit_id)
    REFERENCES business_units (organization_id, id) ON DELETE CASCADE,
  DROP CONSTRAINT orders_status_check,
  ADD CONSTRAINT orders_status_check CHECK (status IN ('baru', 'diproses', 'selesai', 'dibatalkan'));

UPDATE orders o
SET business_unit_id = (SELECT min(i.business_unit_id) FROM order_items i WHERE i.order_id = o.id)
WHERE o.partner_id IS NULL;

-- Invoice berasal dari tepat satu sumber: quotation atau Order.
ALTER TABLE invoices
  ALTER COLUMN quotation_id DROP NOT NULL,
  ADD COLUMN order_id BIGINT UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
  ADD CONSTRAINT invoices_source_chk CHECK ((quotation_id IS NULL) <> (order_id IS NULL)),
  DROP CONSTRAINT invoices_status_check,
  ADD CONSTRAINT invoices_status_check CHECK (status IN ('belum_lunas', 'lunas', 'dibatalkan'));

INSERT INTO invoices (organization_id, order_id, partner_id, customer_id, total)
SELECT organization_id, id, partner_id, customer_id, total FROM orders;

-- Asal persetujuan quotation: langsung oleh customer atau catatan offline.
ALTER TABLE quotations
  ADD COLUMN response_source VARCHAR(20) CHECK (response_source IN ('customer', 'offline')),
  ADD COLUMN response_note TEXT;

ALTER TABLE service_requests
  ADD COLUMN status_changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD CONSTRAINT service_requests_status_check CHECK (status IN ('baru', 'diproses', 'selesai'));
CREATE INDEX service_requests_user_idx ON service_requests (user_id, created_at DESC);

-- Partner hanya melihat pelanggan yang terhubung dengannya.
CREATE TABLE partner_customers (
  organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  partner_id      BIGINT NOT NULL,
  customer_id     BIGINT NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (partner_id, customer_id),
  FOREIGN KEY (organization_id, partner_id) REFERENCES partners (organization_id, id) ON DELETE CASCADE,
  FOREIGN KEY (organization_id, customer_id) REFERENCES customers (organization_id, id) ON DELETE CASCADE
);

INSERT INTO partner_customers (organization_id, partner_id, customer_id)
SELECT organization_id, partner_id, customer_id FROM quotations WHERE partner_id IS NOT NULL
UNION
SELECT organization_id, partner_id, customer_id FROM orders WHERE partner_id IS NOT NULL;

CREATE VIEW product_details AS
SELECT p.id, p.organization_id, p.business_unit_id, p.partner_id, p.type, p.name,
       p.category, p.description, p.price, p.is_active, p.created_at, p.updated_at,
       p.stock, b.name AS business_unit_name, b.is_active AS business_unit_is_active,
       pa.business_name AS partner_name, pa.status AS partner_status
FROM products p
LEFT JOIN business_units b ON b.id = p.business_unit_id
LEFT JOIN partners pa ON pa.id = p.partner_id;

CREATE VIEW quotation_details AS
SELECT q.id, q.organization_id, q.partner_id, q.customer_id, q.status, q.notes,
       q.total, q.responded_at, q.created_at, q.response_source, q.response_note,
       c.name AS customer_name, pa.business_name AS partner_name, i.id AS invoice_id
FROM quotations q
JOIN customers c ON c.id = q.customer_id
LEFT JOIN partners pa ON pa.id = q.partner_id
LEFT JOIN invoices i ON i.quotation_id = q.id;

CREATE VIEW invoice_details AS
SELECT i.id, i.organization_id, i.quotation_id, i.order_id, i.partner_id, i.customer_id,
       i.total, i.status, i.paid_at, i.created_at, c.name AS customer_name,
       pa.business_name AS partner_name
FROM invoices i
JOIN customers c ON c.id = i.customer_id
LEFT JOIN partners pa ON pa.id = i.partner_id;

CREATE VIEW order_details AS
SELECT o.id, o.organization_id, o.partner_id, o.business_unit_id, o.customer_id, o.status,
       o.total, o.status_changed_at, o.created_at, c.name AS customer_name,
       pa.business_name AS partner_name, b.name AS business_unit_name,
       i.id AS invoice_id, i.status AS invoice_status
FROM orders o
JOIN customers c ON c.id = o.customer_id
LEFT JOIN partners pa ON pa.id = o.partner_id
LEFT JOIN business_units b ON b.id = o.business_unit_id
LEFT JOIN invoices i ON i.order_id = o.id;

CREATE VIEW service_request_details AS
SELECT s.id, s.organization_id, s.business_unit_id, s.user_id, s.contact_name,
       s.contact_email, s.contact_phone, s.subject, s.message, s.status,
       s.created_at, s.status_changed_at, b.name AS business_unit_name
FROM service_requests s
JOIN business_units b ON b.id = s.business_unit_id;
