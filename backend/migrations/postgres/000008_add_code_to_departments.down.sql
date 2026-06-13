DROP INDEX IF EXISTS idx_departments_code;
ALTER TABLE departments DROP COLUMN IF EXISTS code;
