CREATE TABLE projects (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Fast lookup of projects by owner
CREATE INDEX idx_projects_owner_id ON projects(owner_id);
