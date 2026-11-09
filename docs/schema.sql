-- =============================================================================
-- Schema database Araya Business Hub (M1-M5), versi migrasi 000005.
--
-- File ini gabungan akhir dari:
--   migrations/000001_init_m1.up.sql
--   migrations/000002_permanent_delete.up.sql
--   migrations/000003_business_modules.up.sql
--   migrations/000004_flow_refinements.up.sql
--   migrations/000005_organization_applications.up.sql
-- Sumber kebenaran tetap folder `migrations/`. Kalau ada migrasi baru,
-- perbarui file ini juga.
--
-- Cara pakai: jalankan sekali di database PostgreSQL kosong (mis. SQL Editor
-- Supabase). Akun super admin TIDAK dibuat di sini karena butuh hash bcrypt;
-- buat lewat `go run ./cmd/seed` dari repo backend.
-- =============================================================================

BEGIN;

-- Organisasi (tenant). Setiap organisasi mengelola pengguna, peran, dan
-- divisinya sendiri. Didaftarkan oleh super admin platform.
CREATE TABLE organizations (
  id             BIGSERIAL PRIMARY KEY,
  code           VARCHAR(50) NOT NULL,
  name           VARCHAR(150) NOT NULL,
  is_active      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  deactivated_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX organizations_code_lower_uidx ON organizations (lower(code));

-- Katalog izin global, misalnya m4.invoice.create. Modul `platform` khusus
-- super admin.
CREATE TABLE permissions (
  code        VARCHAR(100) PRIMARY KEY,
  module      VARCHAR(20) NOT NULL,
  description TEXT NOT NULL
);

-- Peran milik satu organisasi. `is_system` = peran administrator bawaan:
-- otomatis memiliki semua izin non-platform dan tidak dapat diubah/dihapus.
CREATE TABLE roles (
  id              BIGSERIAL PRIMARY KEY,
  organization_id BIGINT NOT NULL REFERENCES organizations(id),
  code            VARCHAR(50) NOT NULL,
  name            VARCHAR(100) NOT NULL,
  description     TEXT,
  is_system       BOOLEAN NOT NULL DEFAULT FALSE,
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (organization_id, code),
  UNIQUE (organization_id, id)
);

CREATE TABLE role_permissions (
  role_id         BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_code VARCHAR(100) NOT NULL REFERENCES permissions(code),
  PRIMARY KEY (role_id, permission_code)
);

CREATE TABLE divisions (
  id              BIGSERIAL PRIMARY KEY,
  organization_id BIGINT NOT NULL REFERENCES organizations(id),
  code            VARCHAR(50) NOT NULL,
  name            VARCHAR(100) NOT NULL,
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (organization_id, code),
  UNIQUE (organization_id, id)
);

-- Pengguna organisasi wajib punya organisasi dan peran. Super admin tidak
-- terikat organisasi. Foreign key komposit menjamin peran/divisi berasal dari
-- organisasi yang sama dengan penggunanya. unit_usaha_id milik M2, tanpa FK.
CREATE TABLE users (
  id              BIGSERIAL PRIMARY KEY,
  organization_id BIGINT REFERENCES organizations(id),
  name            VARCHAR(150) NOT NULL,
  email           VARCHAR(255) NOT NULL,
  password_hash   TEXT NOT NULL,
  role_id         BIGINT,
  division_id     BIGINT,
  unit_usaha_id   BIGINT,
  is_super_admin  BOOLEAN NOT NULL DEFAULT FALSE,
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_by      BIGINT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deactivated_at  TIMESTAMPTZ,
  CONSTRAINT users_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
  FOREIGN KEY (organization_id, role_id) REFERENCES roles (organization_id, id),
  FOREIGN KEY (organization_id, division_id) REFERENCES divisions (organization_id, id),
  CONSTRAINT users_scope_chk CHECK (
    (is_super_admin AND organization_id IS NULL AND role_id IS NULL AND division_id IS NULL)
    OR (NOT is_super_admin AND organization_id IS NOT NULL AND role_id IS NOT NULL)
  )
);
CREATE UNIQUE INDEX users_email_lower_uidx ON users (lower(email));
CREATE INDEX users_organization_id_idx ON users (organization_id);

-- Audit log append-only. actor_name, actor_email, dan organization_name
-- disalin saat kejadian, sehingga log tetap terbaca setelah pengguna atau
-- organisasinya dihapus permanen (FK ON DELETE SET NULL, log tidak ikut
-- terhapus). resource_id tanpa FK karena menunjuk ke tabel yang berbeda
-- sesuai kolom resource.
CREATE TABLE audit_logs (
  id                BIGSERIAL PRIMARY KEY,
  organization_id   BIGINT,
  user_id           BIGINT,
  action            VARCHAR(80) NOT NULL,
  resource          VARCHAR(50) NOT NULL,
  resource_id       BIGINT,
  old_value         JSONB,
  new_value         JSONB,
  ip_address        VARCHAR(45),
  user_agent        TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  actor_name        VARCHAR(150),
  actor_email       VARCHAR(255),
  organization_name VARCHAR(150),
  CONSTRAINT audit_logs_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
  CONSTRAINT audit_logs_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL
);
CREATE INDEX audit_logs_org_created_at_idx ON audit_logs (organization_id, created_at DESC);
CREATE INDEX audit_logs_user_id_idx ON audit_logs (user_id);

-- Pengguna beserta kode peran/divisi dan status organisasinya, dipakai
-- untuk login dan daftar pengguna.
CREATE VIEW user_details AS
SELECT u.id, u.organization_id, u.name, u.email, u.password_hash,
       u.role_id, u.division_id, u.unit_usaha_id, u.is_super_admin,
       u.is_active, u.created_at,
       r.code      AS role_code,
       r.is_system AS role_is_system,
       d.code      AS division_code,
       o.name      AS organization_name,
       o.is_active AS organization_is_active
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
LEFT JOIN divisions d ON d.id = u.division_id
LEFT JOIN organizations o ON o.id = u.organization_id;

-- Katalog izin M1 (sama dengan cmd/seed/data.go).
INSERT INTO permissions (code, module, description) VALUES
  ('platform.organizations.manage', 'platform', 'Mengelola organisasi (khusus super admin)'),
  ('m1.users.read',       'm1', 'Melihat daftar dan detail pengguna'),
  ('m1.users.manage',     'm1', 'Menambah, mengubah, dan menonaktifkan pengguna'),
  ('m1.roles.manage',     'm1', 'Mengelola peran dan izinnya'),
  ('m1.divisions.manage', 'm1', 'Mengelola divisi'),
  ('m1.audit.read',       'm1', 'Melihat audit log')
ON CONFLICT (code) DO NOTHING;

COMMIT;

-- Tabel modul M2-M5. Semua data milik satu organisasi dan ikut terhapus saat
-- organisasinya dihapus permanen. Foreign key komposit menjamin data yang
-- saling merujuk berasal dari organisasi yang sama. Uang dalam rupiah bulat.

-- M2: unit usaha.
CREATE TABLE business_units (
  id              BIGSERIAL PRIMARY KEY,
  organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  division_id     BIGINT,
  code            VARCHAR(50) NOT NULL,
  name            VARCHAR(150) NOT NULL,
  category        VARCHAR(100) NOT NULL,
  description     TEXT,
  address         TEXT,
  phone           VARCHAR(30),
  email           VARCHAR(255),
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (organization_id, code),
  UNIQUE (organization_id, id),
  FOREIGN KEY (organization_id, division_id) REFERENCES divisions (organization_id, id)
);

-- M2: data yang dialirkan modul lain ke Dashboard Operasional Dasar.
-- `value` adalah akumulasi harian per unit usaha.
CREATE TABLE dashboard_metrics (
  id               BIGSERIAL PRIMARY KEY,
  organization_id  BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  business_unit_id BIGINT NOT NULL,
  source_module    VARCHAR(20) NOT NULL,
  metric_key       VARCHAR(50) NOT NULL,
  label            VARCHAR(100) NOT NULL,
  unit             VARCHAR(20) NOT NULL,
  value            BIGINT NOT NULL,
  period_date      DATE NOT NULL,
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (organization_id, business_unit_id, source_module, metric_key, period_date),
  FOREIGN KEY (organization_id, business_unit_id) REFERENCES business_units (organization_id, id) ON DELETE CASCADE
);

-- M3: partner. Akun login dibuat saat mendaftar agar partner bisa melihat
-- status pendaftarannya; fitur transaksi baru terbuka setelah terverifikasi.
CREATE TABLE partners (
  id              BIGSERIAL PRIMARY KEY,
  organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id         BIGINT UNIQUE REFERENCES users(id) ON DELETE SET NULL,
  business_name   VARCHAR(150) NOT NULL,
  business_type   VARCHAR(100) NOT NULL,
  contact_name    VARCHAR(150) NOT NULL,
  email           VARCHAR(255) NOT NULL,
  phone           VARCHAR(30) NOT NULL,
  address         TEXT NOT NULL,
  description     TEXT,
  status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending', 'terverifikasi', 'ditolak', 'ditangguhkan')),
  status_reason   TEXT,
  verified_at     TIMESTAMPTZ,
  verified_by     BIGINT REFERENCES users(id) ON DELETE SET NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (organization_id, id)
);
CREATE INDEX partners_org_status_idx ON partners (organization_id, status);

-- M4: produk/jasa. partner_id NULL berarti milik organisasi.
CREATE TABLE products (
  id               BIGSERIAL PRIMARY KEY,
  organization_id  BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  business_unit_id BIGINT NOT NULL,
  partner_id       BIGINT,
  type             VARCHAR(10) NOT NULL CHECK (type IN ('produk', 'jasa')),
  name             VARCHAR(150) NOT NULL,
  category         VARCHAR(100) NOT NULL,
  description      TEXT,
  price            BIGINT NOT NULL CHECK (price >= 0),
  is_active        BOOLEAN NOT NULL DEFAULT TRUE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (organization_id, id),
  FOREIGN KEY (organization_id, business_unit_id) REFERENCES business_units (organization_id, id) ON DELETE CASCADE,
  FOREIGN KEY (organization_id, partner_id) REFERENCES partners (organization_id, id)
);
CREATE INDEX products_org_category_idx ON products (organization_id, category);

-- M4: pelanggan. user_id terisi bila pelanggan punya akun login.
CREATE TABLE customers (
  id              BIGSERIAL PRIMARY KEY,
  organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id         BIGINT UNIQUE REFERENCES users(id) ON DELETE SET NULL,
  name            VARCHAR(150) NOT NULL,
  email           VARCHAR(255) NOT NULL,
  phone           VARCHAR(30),
  address         TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (organization_id, id)
);
CREATE UNIQUE INDEX customers_org_email_lower_uidx ON customers (organization_id, lower(email));

CREATE TABLE quotations (
  id              BIGSERIAL PRIMARY KEY,
  organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  partner_id      BIGINT,
  customer_id     BIGINT NOT NULL,
  status          VARCHAR(20) NOT NULL DEFAULT 'menunggu'
                  CHECK (status IN ('menunggu', 'disetujui', 'ditolak')),
  notes           TEXT,
  total           BIGINT NOT NULL CHECK (total >= 0),
  created_by      BIGINT REFERENCES users(id) ON DELETE SET NULL,
  responded_at    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (organization_id, id),
  FOREIGN KEY (organization_id, partner_id) REFERENCES partners (organization_id, id),
  FOREIGN KEY (organization_id, customer_id) REFERENCES customers (organization_id, id)
);

-- Nama dan harga disalin saat transaksi agar riwayat tidak berubah ketika
-- produk diubah kemudian.
CREATE TABLE quotation_items (
  id               BIGSERIAL PRIMARY KEY,
  quotation_id     BIGINT NOT NULL REFERENCES quotations(id) ON DELETE CASCADE,
  product_id       BIGINT REFERENCES products(id) ON DELETE SET NULL,
  business_unit_id BIGINT NOT NULL,
  product_name     VARCHAR(150) NOT NULL,
  price            BIGINT NOT NULL CHECK (price >= 0),
  qty              INTEGER NOT NULL CHECK (qty > 0),
  subtotal         BIGINT NOT NULL CHECK (subtotal >= 0)
);
CREATE INDEX quotation_items_quotation_idx ON quotation_items (quotation_id);

-- Satu quotation hanya boleh menjadi satu invoice.
CREATE TABLE invoices (
  id              BIGSERIAL PRIMARY KEY,
  organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  quotation_id    BIGINT NOT NULL UNIQUE REFERENCES quotations(id) ON DELETE CASCADE,
  partner_id      BIGINT,
  customer_id     BIGINT NOT NULL,
  total           BIGINT NOT NULL CHECK (total >= 0),
  status          VARCHAR(20) NOT NULL DEFAULT 'belum_lunas'
                  CHECK (status IN ('belum_lunas', 'lunas')),
  paid_at         TIMESTAMPTZ,
  created_by      BIGINT REFERENCES users(id) ON DELETE SET NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (organization_id, partner_id) REFERENCES partners (organization_id, id),
  FOREIGN KEY (organization_id, customer_id) REFERENCES customers (organization_id, id)
);

-- Setiap percobaan bayar tercatat, termasuk yang gagal.
CREATE TABLE payments (
  id          BIGSERIAL PRIMARY KEY,
  invoice_id  BIGINT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
  provider    VARCHAR(20) NOT NULL,
  reference   VARCHAR(100) NOT NULL UNIQUE,
  amount      BIGINT NOT NULL CHECK (amount >= 0),
  status      VARCHAR(20) NOT NULL CHECK (status IN ('berhasil', 'gagal')),
  message     TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX payments_invoice_idx ON payments (invoice_id);

CREATE TABLE orders (
  id                BIGSERIAL PRIMARY KEY,
  organization_id   BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  partner_id        BIGINT,
  customer_id       BIGINT NOT NULL,
  status            VARCHAR(20) NOT NULL DEFAULT 'baru'
                    CHECK (status IN ('baru', 'diproses', 'selesai')),
  total             BIGINT NOT NULL CHECK (total >= 0),
  status_changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (organization_id, id),
  FOREIGN KEY (organization_id, partner_id) REFERENCES partners (organization_id, id),
  FOREIGN KEY (organization_id, customer_id) REFERENCES customers (organization_id, id)
);
CREATE INDEX orders_org_created_idx ON orders (organization_id, created_at DESC);

CREATE TABLE order_items (
  id               BIGSERIAL PRIMARY KEY,
  order_id         BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id       BIGINT REFERENCES products(id) ON DELETE SET NULL,
  business_unit_id BIGINT NOT NULL,
  product_name     VARCHAR(150) NOT NULL,
  price            BIGINT NOT NULL CHECK (price >= 0),
  qty              INTEGER NOT NULL CHECK (qty > 0),
  subtotal         BIGINT NOT NULL CHECK (subtotal >= 0)
);
CREATE INDEX order_items_order_idx ON order_items (order_id);

-- M5: keranjang per pengguna; satu produk satu baris.
CREATE TABLE cart_items (
  user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  qty        INTEGER NOT NULL CHECK (qty > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, product_id)
);

-- M5: permintaan layanan yang diteruskan ke unit usaha.
CREATE TABLE service_requests (
  id               BIGSERIAL PRIMARY KEY,
  organization_id  BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  business_unit_id BIGINT NOT NULL,
  user_id          BIGINT REFERENCES users(id) ON DELETE SET NULL,
  contact_name     VARCHAR(150) NOT NULL,
  contact_email    VARCHAR(255) NOT NULL,
  contact_phone    VARCHAR(30) NOT NULL,
  subject          VARCHAR(150) NOT NULL,
  message          TEXT NOT NULL,
  status           VARCHAR(20) NOT NULL DEFAULT 'baru',
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (organization_id, business_unit_id) REFERENCES business_units (organization_id, id) ON DELETE CASCADE
);
CREATE INDEX service_requests_org_unit_idx ON service_requests (organization_id, business_unit_id, created_at DESC);

CREATE VIEW business_unit_details AS
SELECT b.id, b.organization_id, b.division_id, b.code, b.name, b.category,
       b.description, b.address, b.phone, b.email, b.is_active, b.created_at,
       b.updated_at, d.name AS division_name
FROM business_units b
LEFT JOIN divisions d ON d.id = b.division_id;

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

-- Pengajuan organisasi yang ingin bergabung ke platform. Organisasi dan akun
-- administratornya baru dibuat setelah pengajuan disetujui super admin.
CREATE TABLE organization_applications (
  id                BIGSERIAL PRIMARY KEY,
  code              VARCHAR(50) NOT NULL,
  organization_name VARCHAR(150) NOT NULL,
  admin_name        VARCHAR(150) NOT NULL,
  admin_email       VARCHAR(255) NOT NULL,
  password_hash     TEXT NOT NULL,
  phone             VARCHAR(30) NOT NULL,
  description       TEXT,
  status            VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'disetujui', 'ditolak')),
  status_reason     TEXT,
  organization_id   BIGINT REFERENCES organizations(id) ON DELETE SET NULL,
  reviewed_by       BIGINT REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at       TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Kode dan email hanya boleh dipakai satu pengajuan yang masih menunggu.
CREATE UNIQUE INDEX organization_applications_pending_code_uidx
  ON organization_applications (lower(code)) WHERE status = 'pending';
CREATE UNIQUE INDEX organization_applications_pending_email_uidx
  ON organization_applications (lower(admin_email)) WHERE status = 'pending';
CREATE INDEX organization_applications_status_idx
  ON organization_applications (status, created_at DESC);
