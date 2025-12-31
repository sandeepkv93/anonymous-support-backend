CREATE TABLE content_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content_type VARCHAR(20) NOT NULL,
    content_id VARCHAR(100) NOT NULL,
    reason VARCHAR(50) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reports_status ON content_reports(status);
CREATE INDEX idx_reports_content ON content_reports(content_type, content_id);
