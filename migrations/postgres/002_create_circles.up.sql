CREATE TABLE circles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    max_members INT NOT NULL DEFAULT 50,
    member_count INT NOT NULL DEFAULT 0,
    is_private BOOLEAN NOT NULL DEFAULT false,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_circles_category ON circles(category);
CREATE INDEX idx_circles_created_at ON circles(created_at);
