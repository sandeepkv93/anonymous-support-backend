CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email TEXT,
    password_hash TEXT,
    avatar_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMP NOT NULL DEFAULT NOW(),
    is_anonymous BOOLEAN NOT NULL DEFAULT true,
    is_banned BOOLEAN NOT NULL DEFAULT false,
    is_premium BOOLEAN NOT NULL DEFAULT false,
    strength_points INT NOT NULL DEFAULT 0,
    CONSTRAINT username_length CHECK (char_length(username) >= 3)
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_created_at ON users(created_at);
