CREATE TABLE project_public_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    html_template TEXT NOT NULL DEFAULT '',
    access_mode TEXT NOT NULL DEFAULT 'public',
    password_hash TEXT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    updated_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT uq_project_public_pages_project UNIQUE (project_id),
    CONSTRAINT uq_project_public_pages_slug UNIQUE (slug),
    CONSTRAINT chk_project_public_pages_access_mode CHECK (access_mode IN ('public', 'password_protected'))
);

CREATE INDEX idx_project_public_pages_company_id
    ON project_public_pages (company_id);

CREATE INDEX idx_project_public_pages_deleted_at
    ON project_public_pages (deleted_at);
