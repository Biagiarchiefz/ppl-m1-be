ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_created_by_fkey,
  ADD CONSTRAINT users_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id);

ALTER TABLE audit_logs
  DROP CONSTRAINT IF EXISTS audit_logs_organization_id_fkey,
  DROP CONSTRAINT IF EXISTS audit_logs_user_id_fkey,
  DROP COLUMN IF EXISTS organization_name,
  DROP COLUMN IF EXISTS actor_email,
  DROP COLUMN IF EXISTS actor_name;
