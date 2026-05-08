DROP TABLE IF EXISTS project_members;

ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_sprint_size_nonnegative;
ALTER TABLE projects DROP COLUMN IF EXISTS sprint_start_date;
ALTER TABLE projects DROP COLUMN IF EXISTS sprint_size;
