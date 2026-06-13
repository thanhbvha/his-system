ALTER TABLE departments ADD COLUMN IF NOT EXISTS code VARCHAR(50);

-- Provide a default code for any existing records
UPDATE departments SET code = 'DEPT_' || substr(md5(id::text), 1, 6) WHERE code IS NULL;

-- Make it NOT NULL and UNIQUE
ALTER TABLE departments ALTER COLUMN code SET NOT NULL;
CREATE UNIQUE INDEX idx_departments_code ON departments(code);
