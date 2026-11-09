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
