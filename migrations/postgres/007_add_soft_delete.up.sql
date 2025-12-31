-- Add soft delete support to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- Add soft delete support to circles table
ALTER TABLE circles ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_circles_deleted_at ON circles(deleted_at) WHERE deleted_at IS NOT NULL;

-- Add soft delete support to reports table
ALTER TABLE reports ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_reports_deleted_at ON reports(deleted_at) WHERE deleted_at IS NOT NULL;

-- Add comment
COMMENT ON COLUMN users.deleted_at IS 'Timestamp when the user was soft deleted (NULL if not deleted)';
COMMENT ON COLUMN circles.deleted_at IS 'Timestamp when the circle was soft deleted (NULL if not deleted)';
COMMENT ON COLUMN reports.deleted_at IS 'Timestamp when the report was soft deleted (NULL if not deleted)';
