-- Seed Permissions
INSERT INTO permissions (id, resource, action) VALUES
(gen_random_uuid(), 'queue', 'read'),
(gen_random_uuid(), 'queue', 'checkin'),
(gen_random_uuid(), 'queue', 'manage'),
(gen_random_uuid(), 'visit', 'read'),
(gen_random_uuid(), 'visit', 'write'),
(gen_random_uuid(), 'patient', 'read'),
(gen_random_uuid(), 'patient', 'write'),
(gen_random_uuid(), 'admin', 'full_access')
ON CONFLICT (resource, action) DO NOTHING;

-- Map Admin Role (admin gets *)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin' AND p.resource = 'admin' AND p.action = 'full_access'
ON CONFLICT DO NOTHING;

-- Map Receptionist Role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'receptionist' AND (
  (p.resource = 'queue' AND p.action IN ('read', 'checkin')) OR
  (p.resource = 'patient' AND p.action IN ('read', 'write'))
)
ON CONFLICT DO NOTHING;

-- Map Doctor Role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'doctor' AND (
  (p.resource = 'queue' AND p.action IN ('read', 'manage')) OR
  (p.resource = 'visit' AND p.action IN ('read', 'write')) OR
  (p.resource = 'patient' AND p.action = 'read')
)
ON CONFLICT DO NOTHING;

-- Map Nurse Role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'nurse' AND (
  (p.resource = 'queue' AND p.action IN ('read', 'manage')) OR
  (p.resource = 'visit' AND p.action = 'read') OR
  (p.resource = 'patient' AND p.action = 'read')
)
ON CONFLICT DO NOTHING;
