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
