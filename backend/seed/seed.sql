-- Seed data for TaskFlow
-- Test credentials: test@example.com / password123
-- Hash generated with bcrypt cost=12

-- ─── Seed User ────────────────────────────────────────────────────────────────
INSERT INTO users (id, name, email, password)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Test User',
    'test@example.com',
    '$2a$12$U1P1uAz7BYZmYzW2qmFMM.uAeSEpIj73EyOgIuo7KbQpsX9VBdBsi'
)
ON CONFLICT (email) DO NOTHING;

-- ─── Seed Project ─────────────────────────────────────────────────────────────
INSERT INTO projects (id, name, description, owner_id)
VALUES (
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'Website Redesign',
    'Q2 redesign of the marketing site — new homepage, nav, and product pages.',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
)
ON CONFLICT DO NOTHING;

-- ─── Seed Tasks (3 different statuses) ────────────────────────────────────────
INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, created_by, due_date)
VALUES
    (
        'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a31',
        'Design homepage mockup',
        'Create Figma mockups for the new homepage layout',
        'done',
        'high',
        'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        '2026-04-10'
    ),
    (
        'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a32',
        'Implement responsive navigation',
        'Build the responsive top-nav component with mobile hamburger menu',
        'in_progress',
        'medium',
        'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        '2026-04-20'
    ),
    (
        'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33',
        'Set up CI/CD pipeline',
        'Configure GitHub Actions for automated testing and deployment',
        'todo',
        'low',
        'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
        NULL,
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        '2026-04-30'
    )
ON CONFLICT DO NOTHING;
