CREATE TABLE circle_memberships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    UNIQUE(circle_id, user_id)
);

CREATE INDEX idx_memberships_circle ON circle_memberships(circle_id);
CREATE INDEX idx_memberships_user ON circle_memberships(user_id);
