-- Reverse RBAC Seeding

DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE resource IN ('queue', 'visit', 'patient', 'admin')
);

DELETE FROM permissions
WHERE resource IN ('queue', 'visit', 'patient', 'admin');
