INSERT INTO roles (company_id, name, description, is_system) VALUES
(NULL, 'Owner', 'Full organization access', TRUE),
(NULL, 'Admin', 'Administrative access without platform ownership', TRUE),
(NULL, 'Member', 'Standard project contributor', TRUE),
(NULL, 'Client', 'Client-facing access limited to project ticket intake', TRUE),
(NULL, 'Guest', 'Read-only project access', TRUE)
ON CONFLICT (company_id, name) DO UPDATE SET
description = EXCLUDED.description,
is_system = TRUE;

DELETE FROM role_permissions
WHERE role_id IN (
    SELECT id FROM roles WHERE company_id IS NULL AND name IN ('Owner', 'Admin', 'Member', 'Client', 'Guest')
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.company_id IS NULL AND r.name = 'Owner'
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
WHERE r.company_id IS NULL AND r.name = 'Admin'
AND p.name IN (
'organization:view','organization:update','member:invite','member:update','member:remove','role:manage',
'project:create','project:view','project:update','project:delete','task:create','task:view','task:update','task:delete',
'file:upload','file:view','file:delete','webhook:manage','billing:manage','apikey:manage','audit:view','chat:use'
) ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
WHERE r.company_id IS NULL AND r.name = 'Member'
AND p.name IN ('organization:view','project:create','project:view','project:update','task:create','task:view','task:update','file:upload','file:view','chat:use')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
WHERE r.company_id IS NULL AND r.name = 'Client'
AND p.name IN ('project:view','task:create','task:view')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
WHERE r.company_id IS NULL AND r.name = 'Guest'
AND p.name IN ('organization:view','project:view','task:view','file:view')
ON CONFLICT (role_id, permission_id) DO NOTHING;
