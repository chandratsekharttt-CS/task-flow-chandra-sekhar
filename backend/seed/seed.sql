-- TaskFlow Seed Data
-- Test user: testuser@taskflow.com / password123
--
-- NOTE: This file documents the seed data structure.
-- The Go application handles seeding with proper bcrypt hashing (SEED_DB=true).
-- You can also run this manually with psql after replacing the password hash.

-- Test User (password: "password123", bcrypt cost 12)
-- The hash below is a placeholder — use the Go seeder for correct hashing.
INSERT INTO users (id, name, email, password)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Test User',
    'testuser@taskflow.com',
    '$2a$12$PLACEHOLDER_HASH_USE_GO_SEEDER'
) ON CONFLICT (id) DO NOTHING;

-- Sample Project
INSERT INTO projects (id, name, description, owner_id)
VALUES (
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'Sample Project',
    'A sample project to demonstrate TaskFlow features.',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
) ON CONFLICT (id) DO NOTHING;

-- Task 1: Todo
INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, created_by, due_date)
VALUES (
    'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01',
    'Design the landing page',
    'Create wireframes and mockups for the main landing page.',
    'todo',
    'high',
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    '2025-04-30'
) ON CONFLICT (id) DO NOTHING;

-- Task 2: In Progress
INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, created_by)
VALUES (
    'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02',
    'Set up CI/CD pipeline',
    'Configure GitHub Actions for automated testing and deployment.',
    'in_progress',
    'medium',
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
) ON CONFLICT (id) DO NOTHING;

-- Task 3: Done
INSERT INTO tasks (id, title, description, status, priority, project_id, created_by)
VALUES (
    'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a03',
    'Write project README',
    'Document setup instructions and API reference.',
    'done',
    'low',
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
) ON CONFLICT (id) DO NOTHING;
