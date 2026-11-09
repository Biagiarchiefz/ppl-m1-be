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

-- Katalog izin global. Tiap modul mendaftarkan kodenya lewat repositori
-- kontrak, misalnya m4.invoice.create. Modul `platform` khusus super admin.
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
-- organisasi yang sama dengan penggunanya.
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
  created_by      BIGINT REFERENCES users(id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deactivated_at  TIMESTAMPTZ,
  FOREIGN KEY (organization_id, role_id) REFERENCES roles (organization_id, id),
  FOREIGN KEY (organization_id, division_id) REFERENCES divisions (organization_id, id),
  CONSTRAINT users_scope_chk CHECK (
    (is_super_admin AND organization_id IS NULL AND role_id IS NULL AND division_id IS NULL)
    OR (NOT is_super_admin AND organization_id IS NOT NULL AND role_id IS NOT NULL)
  )
);
CREATE UNIQUE INDEX users_email_lower_uidx ON users (lower(email));
CREATE INDEX users_organization_id_idx ON users (organization_id);

CREATE TABLE audit_logs (
  id              BIGSERIAL PRIMARY KEY,
  organization_id BIGINT,
  user_id         BIGINT,
  action          VARCHAR(80) NOT NULL,
  resource        VARCHAR(50) NOT NULL,
  resource_id     BIGINT,
  old_value       JSONB,
  new_value       JSONB,
  ip_address      VARCHAR(45),
  user_agent      TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
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
