-- Insert default roles
INSERT INTO roles (name, description) VALUES
    ('admin', 'System Administrator'),
    ('doctor', 'Doctor'),
    ('nurse', 'Nurse'),
    ('receptionist', 'Receptionist'),
    ('pharmacist', 'Pharmacist'),
    ('patient', 'Patient')
ON CONFLICT (name) DO NOTHING;

-- Password: Admin@123
-- Hash: $2a$10$WqiuRa2nRAbooSYc3zeqgun5SCQAzugAE3eUeMVC18kQKO753CYBO
INSERT INTO users (username, password_hash, is_active) VALUES
    ('admin', '$2a$10$WqiuRa2nRAbooSYc3zeqgun5SCQAzugAE3eUeMVC18kQKO753CYBO', true)
ON CONFLICT (username) DO NOTHING;

-- Link admin to admin role
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE u.username = 'admin' AND r.name = 'admin'
ON CONFLICT DO NOTHING;
