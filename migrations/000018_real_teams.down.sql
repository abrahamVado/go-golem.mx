DROP INDEX IF EXISTS idx_projects_team_id;

ALTER TABLE projects
    DROP COLUMN IF EXISTS team_id;

DROP INDEX IF EXISTS idx_team_members_user_id;
DROP TABLE IF EXISTS team_members;

DROP INDEX IF EXISTS idx_teams_company_id;
DROP INDEX IF EXISTS uq_teams_company_slug;
DROP TABLE IF EXISTS teams;
