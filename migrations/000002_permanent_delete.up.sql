-- Hapus permanen pengguna dan organisasi, dengan audit log yang tetap utuh.
-- Audit log menyimpan salinan nama/email pelaku dan nama organisasi saat
-- kejadian, sehingga tetap terbaca setelah datanya dihapus. Foreign key
-- memakai ON DELETE SET NULL: log tidak ikut terhapus.

ALTER TABLE audit_logs
  ADD COLUMN actor_name        VARCHAR(150),
  ADD COLUMN actor_email       VARCHAR(255),
  ADD COLUMN organization_name VARCHAR(150);

UPDATE audit_logs a
SET actor_name = u.name, actor_email = u.email
FROM users u
WHERE u.id = a.user_id;

UPDATE audit_logs a
SET organization_name = o.name
FROM organizations o
WHERE o.id = a.organization_id;

-- Id yang sudah tidak menunjuk ke data mana pun dikosongkan supaya foreign
-- key bisa dipasang.
UPDATE audit_logs a SET user_id = NULL
WHERE user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = a.user_id);

UPDATE audit_logs a SET organization_id = NULL
WHERE organization_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM organizations o WHERE o.id = a.organization_id);

ALTER TABLE audit_logs
  ADD CONSTRAINT audit_logs_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
  ADD CONSTRAINT audit_logs_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL;

-- Menghapus pengguna tidak boleh tertahan oleh akun yang pernah dibuatnya.
ALTER TABLE users
  DROP CONSTRAINT users_created_by_fkey,
  ADD CONSTRAINT users_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
