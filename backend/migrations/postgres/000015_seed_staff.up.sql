-- Seed Departments
INSERT INTO departments (id, code, name, description, is_active) VALUES
(gen_random_uuid(), 'KKB', 'Khoa Khám Bệnh', 'Khoa tiếp nhận và khám bệnh ban đầu', true),
(gen_random_uuid(), 'KCDHA', 'Khoa Chẩn Đoán Hình Ảnh', 'Khoa siêu âm, X-Quang, MRI', true),
(gen_random_uuid(), 'KHSCC', 'Khoa Hồi Sức Cấp Cứu', 'Khoa cấp cứu 24/7', true)
ON CONFLICT DO NOTHING;

-- Seed Staff Profiles for admin
INSERT INTO staff_profiles (id, user_id, full_name, title, department_id, specialty, avatar_url, bio)
SELECT gen_random_uuid(), u.id, 'Quản Trị Hệ Thống', 'Admin', d.id, 'IT', '', 'System Administrator'
FROM users u, departments d
WHERE u.username = 'admin' AND d.code = 'KKB'
ON CONFLICT DO NOTHING;

-- Seed Staff Profiles for letan
INSERT INTO staff_profiles (id, user_id, full_name, title, department_id, specialty, avatar_url, bio)
SELECT gen_random_uuid(), u.id, 'Nguyễn Thu Hà', 'Lễ Tân Tiếp Đón', d.id, 'Lễ Tân', '', 'Lễ tân tiếp đón bệnh nhân tại sảnh chính'
FROM users u, departments d
WHERE u.username = 'letan' AND d.code = 'KKB'
ON CONFLICT DO NOTHING;

-- Seed Staff Profiles for laotu (doctor)
INSERT INTO staff_profiles (id, user_id, full_name, title, department_id, specialty, avatar_url, bio)
SELECT gen_random_uuid(), u.id, 'BS. Lão Tử', 'Bác sĩ Trưởng Khoa', d.id, 'Nội tổng hợp', '', 'Bác sĩ chuyên khoa II nội tổng hợp'
FROM users u, departments d
WHERE u.username = 'laotu' AND d.code = 'KKB'
ON CONFLICT DO NOTHING;
