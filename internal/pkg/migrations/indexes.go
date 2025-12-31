package migrations

// Recommended indexes for production performance
// Run these after initial schema creation

const PostgresIndexes = `
-- Users table indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_last_active ON users(last_active_at DESC);

-- Circles table indexes
CREATE INDEX IF NOT EXISTS idx_circles_category ON circles(category);
CREATE INDEX IF NOT EXISTS idx_circles_created_at ON circles(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_circles_is_private ON circles(is_private);

-- Circle memberships indexes
CREATE INDEX IF NOT EXISTS idx_circle_memberships_circle_id ON circle_memberships(circle_id);
CREATE INDEX IF NOT EXISTS idx_circle_memberships_user_id ON circle_memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_circle_memberships_joined_at ON circle_memberships(joined_at DESC);

-- Moderation indexes
CREATE INDEX IF NOT EXISTS idx_content_reports_status ON content_reports(status);
CREATE INDEX IF NOT EXISTS idx_content_reports_created_at ON content_reports(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_content_reports_reporter_id ON content_reports(reporter_id);

-- Audit logs indexes
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_target_id ON audit_logs(target_id);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_users_active_role ON users(last_active_at DESC, role) WHERE NOT is_banned;
CREATE INDEX IF NOT EXISTS idx_circles_category_members ON circles(category, member_count DESC) WHERE NOT is_private;
`

const MongoDBIndexes = `
// MongoDB indexes - run via mongo shell or driver
// Posts collection
db.posts.createIndex({ "user_id": 1 });
db.posts.createIndex({ "created_at": -1 });
db.posts.createIndex({ "type": 1, "created_at": -1 });
db.posts.createIndex({ "categories": 1, "created_at": -1 });
db.posts.createIndex({ "urgency_level": -1, "created_at": -1 });
db.posts.createIndex({ "circle_id": 1, "created_at": -1 });
db.posts.createIndex({ "visibility": 1, "is_moderated": 1, "created_at": -1 });
db.posts.createIndex({ "deleted_at": 1 }, { sparse: true }); // Soft delete

// TTL index for auto-expiring old posts (optional - 90 days)
db.posts.createIndex({ "created_at": 1 }, { expireAfterSeconds: 7776000 });

// Support responses collection
db.support_responses.createIndex({ "post_id": 1, "created_at": -1 });
db.support_responses.createIndex({ "user_id": 1, "created_at": -1 });
db.support_responses.createIndex({ "type": 1 });

// User analytics collection
db.user_analytics.createIndex({ "user_id": 1 }, { unique: true });
db.user_analytics.createIndex({ "current_streak": -1 });
db.user_analytics.createIndex({ "total_support_given": -1 });
`
