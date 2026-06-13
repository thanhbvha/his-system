INSERT INTO permissions (id, resource, action) VALUES
    (gen_random_uuid(), 'users', 'read'),
    (gen_random_uuid(), 'users', 'write'),
    (gen_random_uuid(), 'users', 'delete'),
    (gen_random_uuid(), 'roles', 'read'),
    (gen_random_uuid(), 'roles', 'write'),
    (gen_random_uuid(), 'departments', 'read'),
    (gen_random_uuid(), 'departments', 'write'),
    (gen_random_uuid(), 'patients', 'read'),
    (gen_random_uuid(), 'patients', 'write'),
    (gen_random_uuid(), 'patients', 'delete'),
    (gen_random_uuid(), 'appointments', 'read'),
    (gen_random_uuid(), 'appointments', 'write'),
    (gen_random_uuid(), 'billing', 'read'),
    (gen_random_uuid(), 'billing', 'write')
ON CONFLICT (resource, action) DO NOTHING;
