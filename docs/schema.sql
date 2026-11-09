-- =============================================================================
-- Schema database Modul M1 (Araya Business Hub), versi migrasi 000002.
--
-- File ini gabungan akhir dari:
--   migrations/000001_init_m1.up.sql
--   migrations/000002_permanent_delete.up.sql
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
