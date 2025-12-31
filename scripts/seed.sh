#!/bin/bash

set -e

echo "=== Seeding Database with Sample Data ==="
echo

echo "Creating sample users..."
psql "postgresql://support_user:support_pass@localhost:5432/support_db?sslmode=disable" <<EOF
INSERT INTO users (id, username, avatar_id, is_anonymous, strength_points) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'john_doe', 1, true, 150),
    ('550e8400-e29b-41d4-a716-446655440002', 'jane_smith', 2, true, 200),
    ('550e8400-e29b-41d4-a716-446655440003', 'mike_wilson', 3, true, 75)
ON CONFLICT DO NOTHING;
EOF
echo "✓ Sample users created"
echo

echo "Creating sample circles..."
psql "postgresql://support_user:support_pass@localhost:5432/support_db?sslmode=disable" <<EOF
INSERT INTO circles (id, name, description, category, max_members, member_count, created_by) VALUES
    ('660e8400-e29b-41d4-a716-446655440001', 'Alcohol Recovery', 'Support for alcohol addiction recovery', 'alcohol', 100, 0, '550e8400-e29b-41d4-a716-446655440001'),
    ('660e8400-e29b-41d4-a716-446655440002', 'Nicotine Free', 'Quit smoking and vaping together', 'nicotine', 50, 0, '550e8400-e29b-41d4-a716-446655440002')
ON CONFLICT DO NOTHING;
EOF
echo "✓ Sample circles created"
echo

echo "=== Seed Complete ==="
echo
