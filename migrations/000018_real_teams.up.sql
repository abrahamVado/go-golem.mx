CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX uq_teams_company_slug
    ON teams (company_id, slug)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_teams_company_id
    ON teams (company_id)
    WHERE deleted_at IS NULL;

CREATE TABLE team_members (
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    added_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    PRIMARY KEY (team_id, user_id)
);

CREATE INDEX idx_team_members_user_id
    ON team_members (user_id)
    WHERE deleted_at IS NULL;

INSERT INTO teams (id, company_id, name, slug, created_by_user_id, created_at, updated_at)
SELECT
    gen_random_uuid(),
    c.id,
    c.name,
    COALESCE(NULLIF(c.slug, ''), 'workspace'),
    (
        SELECT u.id
        FROM users u
        WHERE u.company_id = c.id AND u.deleted_at IS NULL
        ORDER BY u.created_at ASC
        LIMIT 1
    ),
    NOW(),
    NOW()
FROM companies c
WHERE c.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM teams t
      WHERE t.company_id = c.id AND t.deleted_at IS NULL
  );

INSERT INTO team_members (team_id, user_id, added_by_user_id, created_at, updated_at, deleted_at)
SELECT
    t.id,
    u.id,
    t.created_by_user_id,
    NOW(),
    NOW(),
    NULL
FROM users u
JOIN teams t
    ON t.company_id = u.company_id
   AND t.deleted_at IS NULL
WHERE u.deleted_at IS NULL
  AND t.slug = (
      SELECT MIN(t2.slug)
      FROM teams t2
      WHERE t2.company_id = u.company_id AND t2.deleted_at IS NULL
  )
ON CONFLICT (team_id, user_id)
DO UPDATE SET
    deleted_at = NULL,
    updated_at = NOW();

ALTER TABLE projects
    ADD COLUMN team_id UUID NULL REFERENCES teams(id) ON DELETE RESTRICT;

UPDATE projects p
SET team_id = team_ref.id
FROM (
    SELECT DISTINCT ON (t.company_id)
        t.company_id,
        t.id
    FROM teams t
    WHERE t.deleted_at IS NULL
    ORDER BY t.company_id, t.created_at ASC, t.id ASC
) AS team_ref
WHERE p.company_id = team_ref.company_id
  AND p.team_id IS NULL;

ALTER TABLE projects
    ALTER COLUMN team_id SET NOT NULL;

CREATE INDEX idx_projects_team_id
    ON projects (team_id)
    WHERE deleted_at IS NULL;
