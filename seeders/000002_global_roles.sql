INSERT INTO roles (organization_id, name, description, is_system) VALUES
(NULL, 'Owner', 'Full organization access', TRUE),
(NULL, 'Admin', 'Administrative access without platform ownership', TRUE),
(NULL, 'Member', 'Standard project contributor', TRUE),
(NULL, 'Guest', 'Read-only project access', TRUE)
ON DUPLICATE KEY UPDATE description = VALUES(description), is_system = TRUE;

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.organization_id IS NULL AND r.name = 'Owner';

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
WHERE r.organization_id IS NULL AND r.name = 'Admin'
AND p.slug IN (
'organization:view','organization:update','member:invite','member:update','member:remove','role:manage',
'project:create','project:view','project:update','project:delete','task:create','task:view','task:update','task:delete',
'file:upload','file:view','file:delete','webhook:manage','billing:manage','apikey:manage','audit:view','chat:use'
);

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
WHERE r.organization_id IS NULL AND r.name = 'Member'
AND p.slug IN ('organization:view','project:create','project:view','project:update','task:create','task:view','task:update','file:upload','file:view','chat:use');

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
WHERE r.organization_id IS NULL AND r.name = 'Guest'
AND p.slug IN ('organization:view','project:view','task:view','file:view');
