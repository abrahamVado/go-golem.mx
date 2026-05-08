ALTER TABLE projects
    ADD COLUMN sprint_size INTEGER NULL,
    ADD COLUMN sprint_start_date DATE NULL;

ALTER TABLE projects
    ADD CONSTRAINT projects_sprint_size_nonnegative
    CHECK (sprint_size IS NULL OR sprint_size > 0);

CREATE TABLE project_members (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member',
    invited_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    PRIMARY KEY (project_id, user_id),
    CHECK (role IN ('admin', 'member'))
);

CREATE INDEX idx_project_members_user_id
    ON project_members (user_id)
    WHERE deleted_at IS NULL;
