-- Remove soft delete support
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE circles DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE reports DROP COLUMN IF EXISTS deleted_at;
